// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

var typeConversionMap = map[string]func(object interface{}) (objectAsType interface{}, ok bool){
	"AddRolesToUserPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(AddRolesToUserPayloadOrError)
		return obj, ok
	},

	"AddUserWalletPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(AddUserWalletPayloadOrError)
		return obj, ok
	},

	"AdmireFeedEventPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(AdmireFeedEventPayloadOrError)
		return obj, ok
	},

	"AuthorizationError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(AuthorizationError)
		return obj, ok
	},

	"CollectionByIdOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(CollectionByIDOrError)
		return obj, ok
	},

	"CollectionTokenByIdOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(CollectionTokenByIDOrError)
		return obj, ok
	},

	"CommentOnFeedEventPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(CommentOnFeedEventPayloadOrError)
		return obj, ok
	},

	"CommunityByAddressOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(CommunityByAddressOrError)
		return obj, ok
	},

	"CreateCollectionPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(CreateCollectionPayloadOrError)
		return obj, ok
	},

	"CreateUserPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(CreateUserPayloadOrError)
		return obj, ok
	},

	"DeepRefreshPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(DeepRefreshPayloadOrError)
		return obj, ok
	},

	"DeleteCollectionPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(DeleteCollectionPayloadOrError)
		return obj, ok
	},

	"Error": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(Error)
		return obj, ok
	},

	"FeedEventByIdOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(FeedEventByIDOrError)
		return obj, ok
	},

	"FeedEventData": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(FeedEventData)
		return obj, ok
	},

	"FeedEventOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(FeedEventOrError)
		return obj, ok
	},

	"FollowUserPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(FollowUserPayloadOrError)
		return obj, ok
	},

	"GalleryUserOrAddress": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(GalleryUserOrAddress)
		return obj, ok
	},

	"GalleryUserOrWallet": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(GalleryUserOrWallet)
		return obj, ok
	},

	"GetAuthNoncePayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(GetAuthNoncePayloadOrError)
		return obj, ok
	},

	"GroupedNotification": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(GroupedNotification)
		return obj, ok
	},

	"Interaction": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(Interaction)
		return obj, ok
	},

	"LoginPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(LoginPayloadOrError)
		return obj, ok
	},

	"Media": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(Media)
		return obj, ok
	},

	"MediaSubtype": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(MediaSubtype)
		return obj, ok
	},

	"Node": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(Node)
		return obj, ok
	},

	"Notification": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(Notification)
		return obj, ok
	},

	"PreverifyEmailPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(PreverifyEmailPayloadOrError)
		return obj, ok
	},

	"RedeemMerchPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(RedeemMerchPayloadOrError)
		return obj, ok
	},

	"RefreshCollectionPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(RefreshCollectionPayloadOrError)
		return obj, ok
	},

	"RefreshContractPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(RefreshContractPayloadOrError)
		return obj, ok
	},

	"RefreshTokenPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(RefreshTokenPayloadOrError)
		return obj, ok
	},

	"RemoveAdmirePayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(RemoveAdmirePayloadOrError)
		return obj, ok
	},

	"RemoveCommentPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(RemoveCommentPayloadOrError)
		return obj, ok
	},

	"RemoveUserWalletsPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(RemoveUserWalletsPayloadOrError)
		return obj, ok
	},

	"ResendVerificationEmailPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(ResendVerificationEmailPayloadOrError)
		return obj, ok
	},

	"RevokeRolesFromUserPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(RevokeRolesFromUserPayloadOrError)
		return obj, ok
	},

	"SetSpamPreferencePayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(SetSpamPreferencePayloadOrError)
		return obj, ok
	},

	"SyncTokensPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(SyncTokensPayloadOrError)
		return obj, ok
	},

	"TokenByIdOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(TokenByIDOrError)
		return obj, ok
	},

	"UnfollowUserPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UnfollowUserPayloadOrError)
		return obj, ok
	},

	"UnsubscribeFromEmailTypePayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UnsubscribeFromEmailTypePayloadOrError)
		return obj, ok
	},

	"UpdateCollectionHiddenPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UpdateCollectionHiddenPayloadOrError)
		return obj, ok
	},

	"UpdateCollectionInfoPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UpdateCollectionInfoPayloadOrError)
		return obj, ok
	},

	"UpdateCollectionTokensPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UpdateCollectionTokensPayloadOrError)
		return obj, ok
	},

	"UpdateEmailNotificationSettingsPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UpdateEmailNotificationSettingsPayloadOrError)
		return obj, ok
	},

	"UpdateEmailPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UpdateEmailPayloadOrError)
		return obj, ok
	},

	"UpdateGalleryCollectionsPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UpdateGalleryCollectionsPayloadOrError)
		return obj, ok
	},

	"UpdateTokenInfoPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UpdateTokenInfoPayloadOrError)
		return obj, ok
	},

	"UpdateUserInfoPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UpdateUserInfoPayloadOrError)
		return obj, ok
	},

	"UserByAddressOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UserByAddressOrError)
		return obj, ok
	},

	"UserByIdOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UserByIDOrError)
		return obj, ok
	},

	"UserByUsernameOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(UserByUsernameOrError)
		return obj, ok
	},

	"VerifyEmailPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(VerifyEmailPayloadOrError)
		return obj, ok
	},

	"ViewGalleryPayloadOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(ViewGalleryPayloadOrError)
		return obj, ok
	},

	"ViewerOrError": func(object interface{}) (interface{}, bool) {
		obj, ok := object.(ViewerOrError)
		return obj, ok
	},
}

func ConvertToModelType(object interface{}, gqlTypeName string) (objectAsType interface{}, ok bool) {
	if conversionFunc, ok := typeConversionMap[gqlTypeName]; ok {
		if convertedObj, ok := conversionFunc(object); ok {
			return convertedObj, true
		}
	}

	return nil, false
}
