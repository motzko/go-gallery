package tokenprocessing

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/mikeydub/go-gallery/env"
	"github.com/mikeydub/go-gallery/service/logger"
	"github.com/mikeydub/go-gallery/service/media"
	"github.com/mikeydub/go-gallery/service/mediamapper"
	"github.com/mikeydub/go-gallery/service/multichain/opensea"
	"github.com/mikeydub/go-gallery/service/multichain/tezos"
	"github.com/mikeydub/go-gallery/util/retry"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/googleapi"

	"cloud.google.com/go/storage"
	"github.com/everFinance/goar"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/mikeydub/go-gallery/service/rpc"
	"github.com/mikeydub/go-gallery/util"
)

func init() {
	env.RegisterValidation("IPFS_URL", "required")
}

var errAlreadyHasMedia = errors.New("token already has preview and thumbnail URLs")

type errUnsupportedURL struct {
	url string
}

type errUnsupportedMediaType struct {
	mediaType persist.MediaType
}

type errNoDataFromReader struct {
	err error
	url string
}

type errNotCacheable struct {
	URL       string
	MediaType persist.MediaType
}

type errInvalidMedia struct {
	err error
	URL string
}

type errNoCachedObjects struct {
	tids persist.TokenIdentifiers
}

type errNoMediaURLs struct {
	metadata persist.TokenMetadata
	tokenURI persist.TokenURI
	tids     persist.TokenIdentifiers
}

func (e errNoDataFromReader) Error() string {
	return fmt.Sprintf("no data from reader: %s (url: %s)", e.err, e.url)
}

func (e errNotCacheable) Error() string {
	return fmt.Sprintf("unsupported media for caching: %s (mediaURL: %s)", e.MediaType, e.URL)
}

func (e errInvalidMedia) Error() string {
	return fmt.Sprintf("invalid media: %s (url: %s)", e.err, e.URL)
}

func (e errNoMediaURLs) Error() string {
	return fmt.Sprintf("no media URLs found in metadata: %s (metadata: %+v, tokenURI: %s)", e.tids, e.metadata, e.tokenURI)
}

func (e errNoCachedObjects) Error() string {
	return fmt.Sprintf("no cached objects found for token identifiers: %s", e.tids)
}

// MediaProcessingError is an error that occurs when handling media processiing for a token.
type MediaProcessingError struct {
	Chain           persist.Chain
	ContractAddress persist.Address
	TokenID         persist.TokenID
	AnimationError  error
	ImageError      error
}

func (m MediaProcessingError) Error() string {
	return fmt.Sprintf("error with media for token(chain=%d,contractAddress=%s,tokenID=%s): animationErr=%s;imageErr=%s", m.Chain, m.ContractAddress, m.TokenID, m.AnimationError, m.ImageError)
}

type cachePipelineMetadata struct {
	ContentHeaderValueRetrieval  *persist.PipelineStepStatus
	ReaderRetrieval              *persist.PipelineStepStatus
	OpenseaFallback              *persist.PipelineStepStatus
	DetermineMediaTypeWithReader *persist.PipelineStepStatus
	AnimationGzip                *persist.PipelineStepStatus
	StoreGCP                     *persist.PipelineStepStatus
	ThumbnailGCP                 *persist.PipelineStepStatus
	LiveRenderGCP                *persist.PipelineStepStatus
}

// cacheObjectsForMetadata uses a metadata map to generate media content and cache resized versions of the media content.
func cacheObjectsForMetadata(pCtx context.Context, metadata persist.TokenMetadata, contractAddress persist.Address, tokenID persist.TokenID, tokenURI persist.TokenURI, chain persist.Chain, ipfsClient *shell.Shell, arweaveClient *goar.Client, storageClient *storage.Client, tokenBucket string, pMeta *persist.PipelineMetadata) ([]CachedMediaObject, error) {

	pCtx = logger.NewContextWithFields(pCtx, logrus.Fields{
		"contractAddress": contractAddress,
		"tokenID":         tokenID,
		"chain":           chain,
	})
	tids := persist.NewTokenIdentifiers(contractAddress, tokenID, chain)

	imgURL, animURL, err := findImageAndAnimationURLs(pCtx, tokenID, contractAddress, chain, metadata, tokenURI, true, pMeta)
	if err != nil {
		return nil, err
	}

	pCtx = logger.NewContextWithFields(pCtx, logrus.Fields{
		"imgURL":  imgURL,
		"animURL": animURL,
	})

	logger.For(pCtx).Infof("found media URLs")

	var (
		imgCh, animCh         chan cacheResult
		imgResult, animResult cacheResult
	)

	if animURL != "" {
		subMeta := &cachePipelineMetadata{
			ContentHeaderValueRetrieval:  &pMeta.AnimationContentHeaderValueRetrieval,
			ReaderRetrieval:              &pMeta.AnimationReaderRetrieval,
			OpenseaFallback:              &pMeta.AnimationOpenseaFallback,
			DetermineMediaTypeWithReader: &pMeta.AnimationDetermineMediaTypeWithReader,
			AnimationGzip:                &pMeta.AnimationAnimationGzip,
			StoreGCP:                     &pMeta.AnimationStoreGCP,
			ThumbnailGCP:                 &pMeta.AnimationThumbnailGCP,
			LiveRenderGCP:                &pMeta.AnimationLiveRenderGCP,
		}
		animCh = asyncCacheObjectsForURL(pCtx, tids, storageClient, arweaveClient, ipfsClient, ObjectTypeAnimation, animURL, tokenBucket, subMeta)
	}
	if imgURL != "" {
		subMeta := &cachePipelineMetadata{
			ContentHeaderValueRetrieval:  &pMeta.ImageContentHeaderValueRetrieval,
			ReaderRetrieval:              &pMeta.ImageReaderRetrieval,
			OpenseaFallback:              &pMeta.ImageOpenseaFallback,
			DetermineMediaTypeWithReader: &pMeta.ImageDetermineMediaTypeWithReader,
			AnimationGzip:                &pMeta.ImageAnimationGzip,
			StoreGCP:                     &pMeta.ImageStoreGCP,
			ThumbnailGCP:                 &pMeta.ImageThumbnailGCP,
			LiveRenderGCP:                &pMeta.ImageLiveRenderGCP,
		}
		imgCh = asyncCacheObjectsForURL(pCtx, tids, storageClient, arweaveClient, ipfsClient, ObjectTypeImage, imgURL, tokenBucket, subMeta)
	}

	if animCh != nil {
		animResult = <-animCh
	}
	if imgCh != nil {
		imgResult = <-imgCh
	}

	objects := append(animResult.cachedObjects, imgResult.cachedObjects...)

	// the media type is not cacheable but is valid
	if notCacheableErr, ok := animResult.err.(errNotCacheable); ok {
		return nil, notCacheableErr
	} else if notCacheableErr, ok := imgResult.err.(errNotCacheable); ok {
		return nil, notCacheableErr
	}

	// neither download worked, unexpectedly
	if (animCh == nil || (animResult.err != nil && len(animResult.cachedObjects) == 0)) && (imgCh == nil || (imgResult.err != nil && len(imgResult.cachedObjects) == 0)) {
		defer persist.TrackStepStatus(&pMeta.NothingCachedWithErrors)()
		if animCh != nil {
			if _, ok := animResult.err.(errInvalidMedia); ok {
				return nil, MediaProcessingError{
					Chain:           chain,
					ContractAddress: contractAddress,
					TokenID:         tokenID,
					AnimationError:  animResult.err,
					ImageError:      imgResult.err,
				}
			}
		}
		if imgCh != nil {
			if _, ok := imgResult.err.(errInvalidMedia); ok {
				return nil, MediaProcessingError{
					Chain:           chain,
					ContractAddress: contractAddress,
					TokenID:         tokenID,
					AnimationError:  animResult.err,
					ImageError:      imgResult.err,
				}
			}
		}

		return nil, util.MultiErr{animResult.err, imgResult.err}
	}

	// this should never be true, something is wrong if this is true
	if len(objects) == 0 {
		defer persist.TrackStepStatus(&pMeta.NothingCachedWithoutErrors)()
		return nil, errNoCachedObjects{tids: tids}
	}

	return objects, nil
}

func createRawMedia(pCtx context.Context, tids persist.TokenIdentifiers, mediaType persist.MediaType, tokenBucket, animURL, imgURL string, objects []CachedMediaObject) persist.Media {
	switch mediaType {
	case persist.MediaTypeHTML:
		return getHTMLMedia(pCtx, tids, tokenBucket, animURL, imgURL, objects)
	default:
		panic(fmt.Sprintf("media type %s should be cached", mediaType))
	}

}

func createMediaFromCachedObjects(ctx context.Context, tokenBucket string, objects []CachedMediaObject) persist.Media {
	var primaryObject CachedMediaObject
	for _, obj := range objects {
		switch obj.ObjectType {
		case ObjectTypeAnimation:
			// if we receive an animation, that takes top priority and will be the primary object
			primaryObject = obj
			break
		case ObjectTypeImage, ObjectTypeSVG:
			// if we don't have an animation, an image like object will be the primary object
			primaryObject = obj
		}
	}

	var thumbnailObject *CachedMediaObject
	var liveRenderObject *CachedMediaObject
	if primaryObject.ObjectType == ObjectTypeAnimation {
		// animations should have a thumbnail that could be an image or svg or thumbnail
		// thumbnail take top priority, then the other image types that could have been cached
		for _, obj := range objects {
			if obj.ObjectType == ObjectTypeImage || obj.ObjectType == ObjectTypeSVG {
				thumbnailObject = &obj
			} else if obj.ObjectType == ObjectTypeThumbnail {
				thumbnailObject = &obj
				break
			}
		}
	} else {
		// it's not an animation, meaning its image like, so we only use a thumbnail when there explicitly is one
		for _, obj := range objects {
			if obj.ObjectType == ObjectTypeThumbnail {
				thumbnailObject = &obj
				break
			}
		}
	}

	// live render can apply to any media type, if one explicitly exists, use it
	for _, obj := range objects {
		if obj.ObjectType == ObjectTypeLiveRender {
			liveRenderObject = &obj
			break
		}
	}

	result := persist.Media{
		MediaURL:  persist.NullString(primaryObject.StorageURL(tokenBucket)),
		MediaType: primaryObject.MediaType,
	}

	if thumbnailObject != nil {
		result.ThumbnailURL = persist.NullString(thumbnailObject.StorageURL(tokenBucket))
	}

	if liveRenderObject != nil {
		result.LivePreviewURL = persist.NullString(liveRenderObject.StorageURL(tokenBucket))
	}

	var err error
	switch result.MediaType {
	case persist.MediaTypeSVG:
		result.Dimensions, err = getSvgDimensions(ctx, result.MediaURL.String())
	default:
		result.Dimensions, err = getMediaDimensions(ctx, result.MediaURL.String())
	}

	if err != nil {
		logger.For(ctx).WithError(err).Error("failed to get dimensions for media")
	}

	return result
}

type cacheResult struct {
	cachedObjects []CachedMediaObject
	err           error
}

func asyncCacheObjectsForURL(ctx context.Context, tids persist.TokenIdentifiers, storageClient *storage.Client, arweaveClient *goar.Client, ipfsClient *shell.Shell, defaultObjectType objectType, mediaURL, bucket string, subMeta *cachePipelineMetadata) chan cacheResult {
	resultCh := make(chan cacheResult)
	ctx = logger.NewContextWithFields(ctx, logrus.Fields{
		"tokenURIType":      persist.TokenURI(mediaURL).Type(),
		"defaultObjectType": defaultObjectType,
		"mediaURL":          mediaURL,
	})

	go func() {
		cachedObjects, err := cacheObjectsFromURL(ctx, tids, mediaURL, defaultObjectType, ipfsClient, arweaveClient, storageClient, bucket, subMeta, false)
		if err == nil {
			resultCh <- cacheResult{cachedObjects, err}
			return
		}

		switch caught := err.(type) {
		case *googleapi.Error:
			panic(fmt.Errorf("googleAPI error %s: %s", caught, err))
		default:
			resultCh <- cacheResult{cachedObjects, err}
		}
	}()

	return resultCh
}

type svgDimensions struct {
	XMLName xml.Name `xml:"svg"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
	Viewbox string   `xml:"viewBox,attr"`
}

func getSvgDimensions(ctx context.Context, url string) (persist.Dimensions, error) {
	buf := &bytes.Buffer{}
	if strings.HasPrefix(url, "http") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return persist.Dimensions{}, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return persist.Dimensions{}, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return persist.Dimensions{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		_, err = io.Copy(buf, resp.Body)
		if err != nil {
			return persist.Dimensions{}, err
		}
	} else {
		buf = bytes.NewBufferString(url)
	}

	if bytes.HasSuffix(buf.Bytes(), []byte(`<!-- Generated by SVGo -->`)) {
		buf = bytes.NewBuffer(bytes.TrimSuffix(buf.Bytes(), []byte(`<!-- Generated by SVGo -->`)))
	}

	var s svgDimensions
	if err := xml.NewDecoder(buf).Decode(&s); err != nil {
		return persist.Dimensions{}, err
	}

	if (s.Width == "" || s.Height == "") && s.Viewbox == "" {
		return persist.Dimensions{}, fmt.Errorf("no dimensions found for %s", url)
	}

	if s.Viewbox != "" {
		parts := strings.Split(s.Viewbox, " ")
		if len(parts) != 4 {
			return persist.Dimensions{}, fmt.Errorf("invalid viewbox for %s", url)
		}
		s.Width = parts[2]
		s.Height = parts[3]

	}

	width, err := strconv.Atoi(s.Width)
	if err != nil {
		return persist.Dimensions{}, err
	}

	height, err := strconv.Atoi(s.Height)
	if err != nil {
		return persist.Dimensions{}, err
	}

	return persist.Dimensions{
		Width:  width,
		Height: height,
	}, nil
}

func getHTMLMedia(pCtx context.Context, tids persist.TokenIdentifiers, tokenBucket, vURL, imgURL string, cachedObjects []CachedMediaObject) persist.Media {
	res := persist.Media{
		MediaType: persist.MediaTypeHTML,
	}

	if vURL != "" {
		logger.For(pCtx).Infof("using vURL for %s: %s", tids, vURL)
		res.MediaURL = persist.NullString(vURL)
	} else if imgURL != "" {
		logger.For(pCtx).Infof("using imgURL for %s: %s", tids, imgURL)
		res.MediaURL = persist.NullString(imgURL)
	}
	if len(cachedObjects) > 0 {
		for _, obj := range cachedObjects {
			if obj.ObjectType == ObjectTypeThumbnail {
				res.ThumbnailURL = persist.NullString(obj.StorageURL(tokenBucket))
				break
			} else if obj.ObjectType == ObjectTypeImage {
				res.ThumbnailURL = persist.NullString(obj.StorageURL(tokenBucket))
			}
		}
	}
	res = remapMedia(res)

	dimensions, err := getHTMLDimensions(pCtx, res.MediaURL.String())
	if err != nil {
		logger.For(pCtx).Errorf("failed to get dimensions for %s: %v", tids, err)
	}

	res.Dimensions = dimensions

	return res
}

type iframeDimensions struct {
	XMLName xml.Name `xml:"iframe"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
}

func getHTMLDimensions(ctx context.Context, url string) (persist.Dimensions, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return persist.Dimensions{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return persist.Dimensions{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return persist.Dimensions{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var s iframeDimensions
	if err := xml.NewDecoder(resp.Body).Decode(&s); err != nil {
		return persist.Dimensions{}, err
	}

	if s.Width == "" || s.Height == "" {
		return persist.Dimensions{}, fmt.Errorf("no dimensions found for %s", url)
	}

	width, err := strconv.Atoi(s.Width)
	if err != nil {
		return persist.Dimensions{}, err
	}

	height, err := strconv.Atoi(s.Height)
	if err != nil {
		return persist.Dimensions{}, err
	}

	return persist.Dimensions{
		Width:  width,
		Height: height,
	}, nil

}

func remapPaths(mediaURL string) string {
	switch persist.TokenURI(mediaURL).Type() {
	case persist.URITypeIPFS, persist.URITypeIPFSAPI:
		path := util.GetURIPath(mediaURL, false)
		return fmt.Sprintf("%s/ipfs/%s", env.GetString("IPFS_URL"), path)
	case persist.URITypeArweave:
		path := util.GetURIPath(mediaURL, false)
		return fmt.Sprintf("https://arweave.net/%s", path)
	default:
		return mediaURL
	}

}

func remapMedia(media persist.Media) persist.Media {
	media.MediaURL = persist.NullString(remapPaths(strings.TrimSpace(media.MediaURL.String())))
	media.ThumbnailURL = persist.NullString(remapPaths(strings.TrimSpace(media.ThumbnailURL.String())))
	if !persist.TokenURI(media.ThumbnailURL).IsRenderable() {
		media.ThumbnailURL = persist.NullString("")
	}
	return media
}

func findImageAndAnimationURLs(ctx context.Context, tokenID persist.TokenID, contractAddress persist.Address, chain persist.Chain, metadata persist.TokenMetadata, tokenURI persist.TokenURI, predict bool, pMeta *persist.PipelineMetadata) (imgURL string, vURL string, err error) {

	defer persist.TrackStepStatus(&pMeta.MediaURLsRetrieval)()

	ctx = logger.NewContextWithFields(ctx, logrus.Fields{"tokenID": tokenID, "contractAddress": contractAddress})
	if metaMedia, ok := metadata["media"].(map[string]any); ok {
		logger.For(ctx).Debugf("found media metadata: %s", metaMedia)
		var mediaType persist.MediaType

		if mime, ok := metaMedia["mimeType"].(string); ok {
			mediaType = media.MediaFromContentType(mime)
		}
		if uri, ok := metaMedia["uri"].(string); ok {
			switch mediaType {
			case persist.MediaTypeImage, persist.MediaTypeSVG, persist.MediaTypeGIF:
				imgURL = uri
			default:
				vURL = uri
			}
		}
	}

	image, anim := KeywordsForToken(tokenID, contractAddress, chain)
	for _, keyword := range image {
		if it, ok := util.GetValueFromMapUnsafe(metadata, keyword, util.DefaultSearchDepth).(string); ok && it != "" {
			logger.For(ctx).Debugf("found initial animation url from '%s': %s", keyword, it)
			vURL = it
			break
		}
	}

	for _, keyword := range anim {
		if it, ok := util.GetValueFromMapUnsafe(metadata, keyword, util.DefaultSearchDepth).(string); ok && it != "" && it != vURL {
			logger.For(ctx).Debugf("found initial image url from '%s': %s", keyword, it)
			imgURL = it
			break
		}
	}

	if imgURL == "" && vURL == "" {
		persist.FailStep(&pMeta.MediaURLsRetrieval)
		return "", "", errNoMediaURLs{metadata: metadata, tokenURI: tokenURI, tids: persist.NewTokenIdentifiers(contractAddress, tokenID, chain)}
	}

	if predict {
		imgURL, vURL = predictTrueURLs(ctx, imgURL, vURL)
	}
	return imgURL, vURL, nil

}

func findNameAndDescription(ctx context.Context, metadata persist.TokenMetadata) (string, string) {
	name, ok := util.GetValueFromMapUnsafe(metadata, "name", util.DefaultSearchDepth).(string)
	if !ok {
		name = ""
	}

	description, ok := util.GetValueFromMapUnsafe(metadata, "description", util.DefaultSearchDepth).(string)
	if !ok {
		description = ""
	}

	return name, description
}

func predictTrueURLs(ctx context.Context, curImg, curV string) (string, string) {
	imgMediaType, _, _, err := media.PredictMediaType(ctx, curImg)
	if err != nil {
		return curImg, curV
	}
	vMediaType, _, _, err := media.PredictMediaType(ctx, curV)
	if err != nil {
		return curImg, curV
	}

	if imgMediaType.IsAnimationLike() && !vMediaType.IsAnimationLike() {
		return curV, curImg
	}

	if !imgMediaType.IsValid() || !vMediaType.IsValid() {
		return curImg, curV
	}

	if imgMediaType.IsMorePriorityThan(vMediaType) {
		return curV, curImg
	}

	return curImg, curV
}

func getThumbnailURL(pCtx context.Context, tokenBucket string, name string, imgURL string, storageClient *storage.Client) string {
	if storageImageURL, err := getMediaServingURL(pCtx, tokenBucket, fmt.Sprintf("image-%s", name), storageClient); err == nil {
		logger.For(pCtx).Infof("found imageURL for thumbnail %s: %s", name, storageImageURL)
		return storageImageURL
	} else if storageImageURL, err = getMediaServingURL(pCtx, tokenBucket, fmt.Sprintf("svg-%s", name), storageClient); err == nil {
		logger.For(pCtx).Infof("found svg for thumbnail %s: %s", name, storageImageURL)
		return storageImageURL
	} else if imgURL != "" && persist.TokenURI(imgURL).IsRenderable() {
		logger.For(pCtx).Infof("using imgURL for thumbnail %s: %s", name, imgURL)
		return imgURL
	} else if storageImageURL, err := getMediaServingURL(pCtx, tokenBucket, fmt.Sprintf("thumbnail-%s", name), storageClient); err == nil {
		logger.For(pCtx).Infof("found thumbnailURL for %s: %s", name, storageImageURL)
		return storageImageURL
	}
	return ""
}

func objectExists(ctx context.Context, client *storage.Client, bucket, fileName string) (bool, error) {
	objHandle := client.Bucket(bucket).Object(fileName)
	_, err := objHandle.Attrs(ctx)
	if err != nil && err != storage.ErrObjectNotExist {
		return false, fmt.Errorf("could not get object attrs for %s: %s", objHandle.ObjectName(), err)
	}
	return err != storage.ErrObjectNotExist, nil
}

func purgeIfExists(ctx context.Context, bucket string, fileName string, client *storage.Client) error {
	exists, err := objectExists(ctx, client, bucket, fileName)
	if err != nil {
		return err
	}
	if exists {
		if err := mediamapper.PurgeImage(ctx, fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, fileName)); err != nil {
			logger.For(ctx).WithError(err).Errorf("could not purge file %s", fileName)
		}
	}

	return nil
}

func persistToStorage(ctx context.Context, client *storage.Client, reader io.Reader, bucket, fileName string, contentType *string, contentLength *int64, metadata map[string]string) error {
	writer := newObjectWriter(ctx, client, bucket, fileName, contentType, contentLength, metadata)
	if err := retryWriteToCloudStorage(ctx, writer, reader); err != nil {
		return fmt.Errorf("could not write to bucket %s for %s: %s", bucket, fileName, err)
	}
	return writer.Close()
}

func retryWriteToCloudStorage(ctx context.Context, writer io.Writer, reader io.Reader) error {
	return retry.RetryFunc(ctx, func(ctx context.Context) error {
		if _, err := io.Copy(writer, reader); err != nil {
			return err
		}
		return nil
	}, storage.ShouldRetry, retry.DefaultRetry)
}

type objectType int

const (
	ObjectTypeUnknown objectType = iota
	ObjectTypeImage
	ObjectTypeAnimation
	ObjectTypeThumbnail
	ObjectTypeLiveRender
	ObjectTypeSVG
)

func (o objectType) String() string {
	switch o {
	case ObjectTypeImage:
		return "image"
	case ObjectTypeAnimation:
		return "animation"
	case ObjectTypeThumbnail:
		return "thumbnail"
	case ObjectTypeLiveRender:
		return "liverender"
	case ObjectTypeSVG:
		return "svg"
	default:
		panic(fmt.Sprintf("unknown object type: %d", o))
	}
}

type CachedMediaObject struct {
	MediaType       persist.MediaType
	TokenID         persist.TokenID
	ContractAddress persist.Address
	Chain           persist.Chain
	ContentType     *string
	ContentLength   *int64
	ObjectType      objectType
}

func (m CachedMediaObject) fileName() string {
	if m.ObjectType.String() == "" || m.TokenID == "" || m.ContractAddress == "" {
		panic(fmt.Sprintf("invalid media object: %+v", m))
	}
	return fmt.Sprintf("%d-%s-%s-%s", m.Chain, m.TokenID, m.ContractAddress, m.ObjectType)
}

func (m CachedMediaObject) StorageURL(tokenBucket string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", tokenBucket, m.fileName())
}

func cacheRawMedia(ctx context.Context, reader io.Reader, tids persist.TokenIdentifiers, mediaType persist.MediaType, contentLength *int64, contentType *string, defaultObjectType objectType, bucket, ogURL string, client *storage.Client, subMeta *cachePipelineMetadata) (CachedMediaObject, error) {
	defer persist.TrackStepStatus(subMeta.StoreGCP)()

	var objectType objectType
	switch mediaType {
	case persist.MediaTypeVideo:
		objectType = ObjectTypeAnimation
	case persist.MediaTypeSVG:
		objectType = ObjectTypeSVG
	case persist.MediaTypeBase64BMP, persist.MediaTypeBase64PNG:
		objectType = ObjectTypeImage
	default:
		objectType = defaultObjectType
	}
	object := CachedMediaObject{
		MediaType:       mediaType,
		TokenID:         tids.TokenID,
		ContractAddress: tids.ContractAddress,
		Chain:           tids.Chain,
		ContentType:     contentType,
		ContentLength:   contentLength,
		ObjectType:      objectType,
	}
	err := persistToStorage(ctx, client, reader, bucket, object.fileName(), object.ContentType, object.ContentLength,
		map[string]string{
			"originalURL": ogURL,
			"mediaType":   mediaType.String(),
		})
	if err != nil {
		persist.FailStep(subMeta.StoreGCP)
		logger.For(ctx).WithError(err).Errorf("could not persist media to storage")
		return CachedMediaObject{}, err
	}
	go purgeIfExists(context.Background(), bucket, object.fileName(), client)
	return object, err
}

func cacheRawAnimationMedia(ctx context.Context, reader io.Reader, tids persist.TokenIdentifiers, mediaType persist.MediaType, bucket, ogURL string, client *storage.Client, subMeta *cachePipelineMetadata) (CachedMediaObject, error) {
	defer persist.TrackStepStatus(subMeta.AnimationGzip)()

	object := CachedMediaObject{
		MediaType:       mediaType,
		TokenID:         tids.TokenID,
		ContractAddress: tids.ContractAddress,
		Chain:           tids.Chain,
		ObjectType:      ObjectTypeAnimation,
	}

	sw := newObjectWriter(ctx, client, bucket, object.fileName(), nil, nil, map[string]string{
		"originalURL": ogURL,
		"mediaType":   mediaType.String(),
	})
	writer := gzip.NewWriter(sw)

	err := retryWriteToCloudStorage(ctx, writer, reader)
	if err != nil {
		persist.FailStep(subMeta.AnimationGzip)
		return CachedMediaObject{}, fmt.Errorf("could not write to bucket %s for %s: %s", bucket, object.fileName(), err)
	}

	if err := writer.Close(); err != nil {
		persist.FailStep(subMeta.AnimationGzip)
		return CachedMediaObject{}, err
	}

	if err := sw.Close(); err != nil {
		persist.FailStep(subMeta.AnimationGzip)
		return CachedMediaObject{}, err
	}

	go purgeIfExists(context.Background(), bucket, object.fileName(), client)
	return object, nil
}

func thumbnailAndCache(ctx context.Context, tids persist.TokenIdentifiers, videoURL, bucket string, client *storage.Client, subMeta *cachePipelineMetadata) (CachedMediaObject, error) {
	defer persist.TrackStepStatus(subMeta.ThumbnailGCP)()
	obj := CachedMediaObject{
		ObjectType:      ObjectTypeThumbnail,
		MediaType:       persist.MediaTypeImage,
		TokenID:         tids.TokenID,
		ContractAddress: tids.ContractAddress,
		Chain:           tids.Chain,
		ContentType:     util.ToPointer("image/png"),
	}

	logger.For(ctx).Infof("caching thumbnail for '%s'", obj.fileName())

	timeBeforeCopy := time.Now()

	sw := newObjectWriter(ctx, client, bucket, obj.fileName(), util.ToPointer("image/jpeg"), nil, map[string]string{
		"thumbnailedURL": videoURL,
	})

	logger.For(ctx).Infof("thumbnailing %s", videoURL)
	if err := thumbnailVideoToWriter(ctx, videoURL, sw); err != nil {
		persist.FailStep(subMeta.ThumbnailGCP)
		return CachedMediaObject{}, fmt.Errorf("could not thumbnail to bucket %s for '%s': %s", bucket, obj.fileName(), err)
	}

	if err := sw.Close(); err != nil {
		persist.FailStep(subMeta.ThumbnailGCP)
		return CachedMediaObject{}, err
	}

	logger.For(ctx).Infof("storage copy took %s", time.Since(timeBeforeCopy))

	go purgeIfExists(context.Background(), bucket, obj.fileName(), client)

	return obj, nil
}

func createLiveRenderAndCache(ctx context.Context, tids persist.TokenIdentifiers, videoURL, bucket string, client *storage.Client, subMeta *cachePipelineMetadata) (CachedMediaObject, error) {

	defer persist.TrackStepStatus(subMeta.LiveRenderGCP)()

	obj := CachedMediaObject{
		ObjectType:      ObjectTypeLiveRender,
		MediaType:       persist.MediaTypeVideo,
		TokenID:         tids.TokenID,
		ContractAddress: tids.ContractAddress,
		Chain:           tids.Chain,
		ContentType:     util.ToPointer("video/mp4"),
	}

	logger.For(ctx).Infof("caching live render media for '%s'", obj.fileName())

	timeBeforeCopy := time.Now()

	sw := newObjectWriter(ctx, client, bucket, obj.fileName(), util.ToPointer("video/mp4"), nil, map[string]string{
		"liveRenderedURL": videoURL,
	})

	logger.For(ctx).Infof("creating live render for %s", videoURL)
	if err := createLiveRenderPreviewVideo(ctx, videoURL, sw); err != nil {
		persist.FailStep(subMeta.LiveRenderGCP)
		return CachedMediaObject{}, fmt.Errorf("could not live render to bucket %s for '%s': %s", bucket, obj.fileName(), err)
	}

	if err := sw.Close(); err != nil {
		persist.FailStep(subMeta.LiveRenderGCP)
		return CachedMediaObject{}, err
	}

	logger.For(ctx).Infof("storage copy took %s", time.Since(timeBeforeCopy))

	go purgeIfExists(context.Background(), bucket, obj.fileName(), client)

	return obj, nil
}

func deleteMedia(ctx context.Context, bucket, fileName string, client *storage.Client) error {
	return client.Bucket(bucket).Object(fileName).Delete(ctx)
}

func getMediaServingURL(pCtx context.Context, bucketID, objectID string, client *storage.Client) (string, error) {
	if exists, err := objectExists(pCtx, client, bucketID, objectID); err != nil || !exists {
		objectName := fmt.Sprintf("/gs/%s/%s", bucketID, objectID)
		return "", fmt.Errorf("failed to check if object %s exists: %s", objectName, err)
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketID, objectID), nil
}

func cacheObjectsFromURL(pCtx context.Context, tids persist.TokenIdentifiers, mediaURL string, defaultObjectType objectType, ipfsClient *shell.Shell, arweaveClient *goar.Client, storageClient *storage.Client, bucket string, subMeta *cachePipelineMetadata, isRecursive bool) ([]CachedMediaObject, error) {

	asURI := persist.TokenURI(mediaURL)
	timeBeforePredict := time.Now()
	mediaType, contentType, contentLength := func() (persist.MediaType, *string, *int64) {
		defer persist.TrackStepStatus(subMeta.ContentHeaderValueRetrieval)()
		mediaType, contentType, contentLength, _ := media.PredictMediaType(pCtx, asURI.String())
		pCtx = logger.NewContextWithFields(pCtx, logrus.Fields{
			"predictedMediaType":   mediaType,
			"predictedContentType": contentType,
		})

		contentLengthStr := "nil"
		if contentLength != nil {
			contentLengthStr = util.InByteSizeFormat(uint64(util.FromPointer(contentLength)))
		}
		pCtx = logger.NewContextWithFields(pCtx, logrus.Fields{
			"contentLength": contentLength,
		})
		logger.For(pCtx).Infof("predicted media type from '%s' as '%s' with length %s in %s", mediaURL, mediaType, contentLengthStr, time.Since(timeBeforePredict))
		return mediaType, contentType, contentLength
	}()

	if mediaType == persist.MediaTypeHTML {
		return nil, errNotCacheable{URL: mediaURL, MediaType: mediaType}
	}

	timeBeforeDataReader := time.Now()
	reader, retryOpensea, err := func() (*util.FileHeaderReader, bool, error) {
		defer persist.TrackStepStatus(subMeta.ReaderRetrieval)()
		reader, err := rpc.GetDataFromURIAsReader(pCtx, asURI, ipfsClient, arweaveClient)
		if err != nil {

			// the reader is and always will be invalid
			switch caught := err.(type) {
			case rpc.ErrHTTP:
				if caught.Status == http.StatusNotFound || caught.Status == http.StatusInternalServerError {
					persist.FailStep(subMeta.ReaderRetrieval)
					return nil, false, errInvalidMedia{URL: mediaURL, err: err}
				}
			case *net.DNSError, *url.Error:
				persist.FailStep(subMeta.ReaderRetrieval)
				return nil, false, errInvalidMedia{URL: mediaURL, err: err}
			}

			if !isRecursive && tids.Chain == persist.ChainETH {
				return nil, true, nil
			}

			// if we're not already recursive, try opensea for ethereum tokens
			persist.FailStep(subMeta.ReaderRetrieval)
			return nil, false, errNoDataFromReader{err: err, url: mediaURL}
		}
		return reader, false, nil
	}()

	if err != nil {
		return nil, err
	}

	if retryOpensea {
		defer persist.TrackStepStatus(subMeta.OpenseaFallback)()
		logger.For(pCtx).Infof("failed to get data from uri '%s' for '%s' because of (err: %s <%T>), trying opensea", mediaURL, tids, err, err)
		// if token is ETH, backup to asking opensea
		assets, err := opensea.FetchAssetsForTokenIdentifiers(pCtx, persist.EthereumAddress(tids.ContractAddress), opensea.TokenID(tids.TokenID.Base10String()))
		if err != nil || len(assets) == 0 {
			// no data from opensea, return error
			return nil, errNoDataFromReader{err: err, url: mediaURL}
		}

		for _, asset := range assets {
			// does this asset have any valid URLs?
			firstNonEmptyURL, ok := util.FindFirst([]string{asset.AnimationURL, asset.ImageURL, asset.ImagePreviewURL, asset.ImageOriginalURL, asset.ImageThumbnailURL}, func(s string) bool {
				return s != ""
			})
			if !ok {
				continue
			}

			reader, err = rpc.GetDataFromURIAsReader(pCtx, persist.TokenURI(firstNonEmptyURL), ipfsClient, arweaveClient)
			if err != nil {
				continue
			}

			logger.For(pCtx).Infof("got reader for %s from opensea in %s (%s)", tids, time.Since(timeBeforeDataReader), firstNonEmptyURL)

			return cacheObjectsFromURL(pCtx, tids, firstNonEmptyURL, defaultObjectType, ipfsClient, arweaveClient, storageClient, bucket, subMeta, true)
		}
	}

	logger.For(pCtx).Infof("got reader for %s in %s", tids, time.Since(timeBeforeDataReader))

	defer reader.Close()

	if !mediaType.IsValid() {
		func() {
			defer persist.TrackStepStatus(subMeta.DetermineMediaTypeWithReader)()
			timeBeforeSniff := time.Now()
			bytesToSniff, err := reader.Headers()
			if err != nil {
				persist.FailStep(subMeta.DetermineMediaTypeWithReader)
				logger.For(pCtx).WithError(err).Errorf("could not get headers for %s", mediaURL)
				return
			}
			contentType = util.ToPointer("")
			mediaType, *contentType = media.SniffMediaType(bytesToSniff)
			logger.For(pCtx).Infof("sniffed media type for %s: %s in %s", truncateString(mediaURL, 50), mediaType, time.Since(timeBeforeSniff))
		}()
	}

	if mediaType == persist.MediaTypeHTML {
		return nil, errNotCacheable{URL: mediaURL, MediaType: mediaType}
	}

	asMb := 0.0
	if contentLength != nil && *contentLength > 0 {
		asMb = float64(*contentLength) / 1024 / 1024
	}

	pCtx = logger.NewContextWithFields(pCtx, logrus.Fields{
		"finalMediaType":   mediaType,
		"finalContentType": contentType,
		"mb":               asMb,
	})

	logger.For(pCtx).Infof("caching %.2f mb of raw media with type '%s' for '%s' at '%s-%s'", asMb, mediaType, mediaURL, defaultObjectType, tids)

	if mediaType == persist.MediaTypeAnimation {
		timeBeforeCache := time.Now()
		obj, err := cacheRawAnimationMedia(pCtx, reader, tids, mediaType, bucket, mediaURL, storageClient, subMeta)
		if err != nil {
			return nil, err
		}
		logger.For(pCtx).Infof("cached animation for %s in %s", tids, time.Since(timeBeforeCache))
		return []CachedMediaObject{obj}, nil
	}

	timeBeforeCache := time.Now()
	obj, err := cacheRawMedia(pCtx, reader, tids, mediaType, contentLength, contentType, defaultObjectType, bucket, mediaURL, storageClient, subMeta)
	if err != nil {
		return nil, err
	}
	logger.For(pCtx).Infof("cached media for %s in %s", tids, time.Since(timeBeforeCache))

	result := []CachedMediaObject{obj}
	if mediaType == persist.MediaTypeVideo {
		videoURL := obj.StorageURL(bucket)
		thumbObj, err := thumbnailAndCache(pCtx, tids, videoURL, bucket, storageClient, subMeta)
		if err != nil {
			logger.For(pCtx).Errorf("could not create thumbnail for %s: %s", tids, err)
		} else {
			result = append(result, thumbObj)
		}

		liveObj, err := createLiveRenderAndCache(pCtx, tids, videoURL, bucket, storageClient, subMeta)
		if err != nil {
			logger.For(pCtx).Errorf("could not create live render for %s: %s", tids, err)
		} else {
			result = append(result, liveObj)
		}

	}

	return result, nil
}

func thumbnailVideoToWriter(ctx context.Context, url string, writer io.Writer) error {
	c := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-i", url, "-ss", "00:00:00.000", "-vframes", "1", "-f", "mjpeg", "pipe:1")
	c.Stderr = os.Stderr
	c.Stdout = writer
	return c.Run()
}

func createLiveRenderPreviewVideo(ctx context.Context, videoURL string, writer io.Writer) error {
	c := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-i", videoURL, "-ss", "00:00:00.000", "-t", "00:00:05.000", "-filter:v", "scale=720:-1", "-movflags", "frag_keyframe+empty_moov", "-c:a", "copy", "-f", "mp4", "pipe:1")
	c.Stderr = os.Stderr
	c.Stdout = writer
	return c.Run()
}

type dimensions struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

type errNoStreams struct {
	url string
	err error
}

func (e errNoStreams) Error() string {
	return fmt.Sprintf("no streams in %s: %s", e.url, e.err)
}

func getMediaDimensions(ctx context.Context, url string) (persist.Dimensions, error) {
	outBuf := &bytes.Buffer{}
	c := exec.CommandContext(ctx, "ffprobe", "-hide_banner", "-loglevel", "error", "-show_streams", url, "-print_format", "json")
	c.Stderr = os.Stderr
	c.Stdout = outBuf
	err := c.Run()
	if err != nil {
		return persist.Dimensions{}, err
	}

	var d dimensions
	err = json.Unmarshal(outBuf.Bytes(), &d)
	if err != nil {
		return persist.Dimensions{}, fmt.Errorf("failed to unmarshal ffprobe output: %w", err)
	}

	if len(d.Streams) == 0 {
		return persist.Dimensions{}, fmt.Errorf("no streams found in ffprobe output: %w", err)
	}

	dims := persist.Dimensions{}

	for _, s := range d.Streams {
		if s.Height == 0 || s.Width == 0 {
			continue
		}
		dims = persist.Dimensions{
			Width:  s.Width,
			Height: s.Height,
		}
		break
	}

	logger.For(ctx).Debugf("got dimensions %+v for %s", dims, url)
	return dims, nil
}

func truncateString(s string, i int) string {
	asRunes := []rune(s)
	if len(asRunes) > i {
		return string(asRunes[:i])
	}
	return s
}

func KeywordsForToken(tokenID persist.TokenID, contract persist.Address, chain persist.Chain) ([]string, []string) {
	switch {
	case tezos.IsHicEtNunc(contract):
		_, anim := chain.BaseKeywords()
		return []string{"artifactUri", "displayUri", "image"}, anim
	case tezos.IsFxHash(contract):
		return []string{"displayUri", "artifactUri", "image", "uri"}, []string{"artifactUri", "displayUri"}
	default:
		return chain.BaseKeywords()
	}
}

func (e errUnsupportedURL) Error() string {
	return fmt.Sprintf("unsupported url %s", e.url)
}

func (e errUnsupportedMediaType) Error() string {
	return fmt.Sprintf("unsupported media type %s", e.mediaType)
}

func newObjectWriter(ctx context.Context, client *storage.Client, bucket, fileName string, contentType *string, contentLength *int64, objMetadata map[string]string) *storage.Writer {
	writer := client.Bucket(bucket).Object(fileName).NewWriter(ctx)
	if contentType != nil {
		writer.ObjectAttrs.ContentType = *contentType
	}
	writer.ObjectAttrs.Metadata = objMetadata
	writer.ObjectAttrs.CacheControl = "no-cache, no-store"
	writer.ChunkSize = 4 * 1024 * 1024 // 4MB
	writer.ChunkRetryDeadline = 1 * time.Minute
	if contentLength != nil {
		cl := *contentLength
		if cl < 4*1024*1024 {
			writer.ChunkSize = int(cl)
		} else if cl > 32*1024*1024 {
			writer.ChunkSize = 8 * 1024 * 1024
		}
	}
	return writer
}