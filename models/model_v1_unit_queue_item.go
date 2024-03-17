/*
Stickerio API

MMO RTS Stickerio game on an API.

API version: 1.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package generated

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the V1UnitQueueItem type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1UnitQueueItem{}

// V1UnitQueueItem struct for V1UnitQueueItem
type V1UnitQueueItem struct {
	Id string `json:"id"`
	QueuedEpoch int64 `json:"queuedEpoch"`
	DurationSec int64 `json:"durationSec"`
	UnitCount int64 `json:"unitCount"`
	UnitType string `json:"unitType"`
}

type _V1UnitQueueItem V1UnitQueueItem

// NewV1UnitQueueItem instantiates a new V1UnitQueueItem object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1UnitQueueItem(id string, queuedEpoch int64, durationSec int64, unitCount int64, unitType string) *V1UnitQueueItem {
	this := V1UnitQueueItem{}
	this.Id = id
	this.QueuedEpoch = queuedEpoch
	this.DurationSec = durationSec
	this.UnitCount = unitCount
	this.UnitType = unitType
	return &this
}

// NewV1UnitQueueItemWithDefaults instantiates a new V1UnitQueueItem object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1UnitQueueItemWithDefaults() *V1UnitQueueItem {
	this := V1UnitQueueItem{}
	return &this
}

// GetId returns the Id field value
func (o *V1UnitQueueItem) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *V1UnitQueueItem) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *V1UnitQueueItem) SetId(v string) {
	o.Id = v
}

// GetQueuedEpoch returns the QueuedEpoch field value
func (o *V1UnitQueueItem) GetQueuedEpoch() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.QueuedEpoch
}

// GetQueuedEpochOk returns a tuple with the QueuedEpoch field value
// and a boolean to check if the value has been set.
func (o *V1UnitQueueItem) GetQueuedEpochOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.QueuedEpoch, true
}

// SetQueuedEpoch sets field value
func (o *V1UnitQueueItem) SetQueuedEpoch(v int64) {
	o.QueuedEpoch = v
}

// GetDurationSec returns the DurationSec field value
func (o *V1UnitQueueItem) GetDurationSec() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.DurationSec
}

// GetDurationSecOk returns a tuple with the DurationSec field value
// and a boolean to check if the value has been set.
func (o *V1UnitQueueItem) GetDurationSecOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.DurationSec, true
}

// SetDurationSec sets field value
func (o *V1UnitQueueItem) SetDurationSec(v int64) {
	o.DurationSec = v
}

// GetUnitCount returns the UnitCount field value
func (o *V1UnitQueueItem) GetUnitCount() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.UnitCount
}

// GetUnitCountOk returns a tuple with the UnitCount field value
// and a boolean to check if the value has been set.
func (o *V1UnitQueueItem) GetUnitCountOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.UnitCount, true
}

// SetUnitCount sets field value
func (o *V1UnitQueueItem) SetUnitCount(v int64) {
	o.UnitCount = v
}

// GetUnitType returns the UnitType field value
func (o *V1UnitQueueItem) GetUnitType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.UnitType
}

// GetUnitTypeOk returns a tuple with the UnitType field value
// and a boolean to check if the value has been set.
func (o *V1UnitQueueItem) GetUnitTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.UnitType, true
}

// SetUnitType sets field value
func (o *V1UnitQueueItem) SetUnitType(v string) {
	o.UnitType = v
}

func (o V1UnitQueueItem) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1UnitQueueItem) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["id"] = o.Id
	toSerialize["queuedEpoch"] = o.QueuedEpoch
	toSerialize["durationSec"] = o.DurationSec
	toSerialize["unitCount"] = o.UnitCount
	toSerialize["unitType"] = o.UnitType
	return toSerialize, nil
}

func (o *V1UnitQueueItem) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"queuedEpoch",
		"durationSec",
		"unitCount",
		"unitType",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varV1UnitQueueItem := _V1UnitQueueItem{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varV1UnitQueueItem)

	if err != nil {
		return err
	}

	*o = V1UnitQueueItem(varV1UnitQueueItem)

	return err
}

type NullableV1UnitQueueItem struct {
	value *V1UnitQueueItem
	isSet bool
}

func (v NullableV1UnitQueueItem) Get() *V1UnitQueueItem {
	return v.value
}

func (v *NullableV1UnitQueueItem) Set(val *V1UnitQueueItem) {
	v.value = val
	v.isSet = true
}

func (v NullableV1UnitQueueItem) IsSet() bool {
	return v.isSet
}

func (v *NullableV1UnitQueueItem) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1UnitQueueItem(val *V1UnitQueueItem) *NullableV1UnitQueueItem {
	return &NullableV1UnitQueueItem{value: val, isSet: true}
}

func (v NullableV1UnitQueueItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1UnitQueueItem) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


