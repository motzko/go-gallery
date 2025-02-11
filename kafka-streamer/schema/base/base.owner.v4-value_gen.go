// Code generated by avrogen. DO NOT EDIT.

package base

import (
	"github.com/heetch/avro/avrotypegen"
)

type Owner struct {
	Nft_id                     string  `json:"nft_id"`
	Owner_address              string  `json:"owner_address"`
	Quantity                   *string `json:"quantity"`
	Collection_id              *string `json:"collection_id"`
	First_acquired_date        *string `json:"first_acquired_date"`
	Last_acquired_date         *string `json:"last_acquired_date"`
	First_acquired_transaction *string `json:"first_acquired_transaction"`
	Last_acquired_transaction  *string `json:"last_acquired_transaction"`
	Minted_to_this_wallet      *bool   `json:"minted_to_this_wallet"`
	Airdropped_to_this_wallet  *bool   `json:"airdropped_to_this_wallet"`
	Sold_to_this_wallet        *bool   `json:"sold_to_this_wallet"`
}

// AvroRecord implements the avro.AvroRecord interface.
func (Owner) AvroRecord() avrotypegen.RecordInfo {
	return avrotypegen.RecordInfo{
		Schema: `{"__fastavro_parsed":true,"__named_schemas":{"com.simplehash.v0.owner":{"fields":[{"name":"nft_id","type":"string"},{"name":"owner_address","type":"string"},{"default":null,"name":"quantity","type":["null","string"]},{"default":null,"name":"collection_id","type":["null","string"]},{"default":null,"name":"first_acquired_date","type":["null","string"]},{"default":null,"name":"last_acquired_date","type":["null","string"]},{"default":null,"name":"first_acquired_transaction","type":["null","string"]},{"default":null,"name":"last_acquired_transaction","type":["null","string"]},{"default":null,"name":"minted_to_this_wallet","type":["null","boolean"]},{"default":null,"name":"airdropped_to_this_wallet","type":["null","boolean"]},{"default":null,"name":"sold_to_this_wallet","type":["null","boolean"]}],"name":"com.simplehash.v0.owner","type":"record"}},"fields":[{"name":"nft_id","type":"string"},{"name":"owner_address","type":"string"},{"default":null,"name":"quantity","type":["null","string"]},{"default":null,"name":"collection_id","type":["null","string"]},{"default":null,"name":"first_acquired_date","type":["null","string"]},{"default":null,"name":"last_acquired_date","type":["null","string"]},{"default":null,"name":"first_acquired_transaction","type":["null","string"]},{"default":null,"name":"last_acquired_transaction","type":["null","string"]},{"default":null,"name":"minted_to_this_wallet","type":["null","boolean"]},{"default":null,"name":"airdropped_to_this_wallet","type":["null","boolean"]},{"default":null,"name":"sold_to_this_wallet","type":["null","boolean"]}],"name":"com.simplehash.v0.owner","type":"record"}`,
		Required: []bool{
			0: true,
			1: true,
		},
	}
}
