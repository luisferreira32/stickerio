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

// checks if the V1Movement type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1Movement{}

// V1Movement struct for V1Movement
type V1Movement struct {
	Id string `json:"id"`
	PlayerID string `json:"playerID"`
	OriginID string `json:"originID"`
	DestinationID string `json:"destinationID"`
	DestinationX int32 `json:"destinationX"`
	DestinationY int32 `json:"destinationY"`
	DepartureEpoch int64 `json:"departureEpoch"`
	Speed float64 `json:"speed"`
	UnitCount map[string]int64 `json:"unitCount"`
	ResourceCount map[string]int64 `json:"resourceCount"`
}

type _V1Movement V1Movement

// NewV1Movement instantiates a new V1Movement object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1Movement(id string, playerID string, originID string, destinationID string, destinationX int32, destinationY int32, departureEpoch int64, speed float64, unitCount map[string]int64, resourceCount map[string]int64) *V1Movement {
	this := V1Movement{}
	this.Id = id
	this.PlayerID = playerID
	this.OriginID = originID
	this.DestinationID = destinationID
	this.DestinationX = destinationX
	this.DestinationY = destinationY
	this.DepartureEpoch = departureEpoch
	this.Speed = speed
	this.UnitCount = unitCount
	this.ResourceCount = resourceCount
	return &this
}

// NewV1MovementWithDefaults instantiates a new V1Movement object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1MovementWithDefaults() *V1Movement {
	this := V1Movement{}
	return &this
}

// GetId returns the Id field value
func (o *V1Movement) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *V1Movement) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *V1Movement) SetId(v string) {
	o.Id = v
}

// GetPlayerID returns the PlayerID field value
func (o *V1Movement) GetPlayerID() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.PlayerID
}

// GetPlayerIDOk returns a tuple with the PlayerID field value
// and a boolean to check if the value has been set.
func (o *V1Movement) GetPlayerIDOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.PlayerID, true
}

// SetPlayerID sets field value
func (o *V1Movement) SetPlayerID(v string) {
	o.PlayerID = v
}

// GetOriginID returns the OriginID field value
func (o *V1Movement) GetOriginID() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.OriginID
}

// GetOriginIDOk returns a tuple with the OriginID field value
// and a boolean to check if the value has been set.
func (o *V1Movement) GetOriginIDOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.OriginID, true
}

// SetOriginID sets field value
func (o *V1Movement) SetOriginID(v string) {
	o.OriginID = v
}

// GetDestinationID returns the DestinationID field value
func (o *V1Movement) GetDestinationID() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.DestinationID
}

// GetDestinationIDOk returns a tuple with the DestinationID field value
// and a boolean to check if the value has been set.
func (o *V1Movement) GetDestinationIDOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.DestinationID, true
}

// SetDestinationID sets field value
func (o *V1Movement) SetDestinationID(v string) {
	o.DestinationID = v
}

// GetDestinationX returns the DestinationX field value
func (o *V1Movement) GetDestinationX() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.DestinationX
}

// GetDestinationXOk returns a tuple with the DestinationX field value
// and a boolean to check if the value has been set.
func (o *V1Movement) GetDestinationXOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.DestinationX, true
}

// SetDestinationX sets field value
func (o *V1Movement) SetDestinationX(v int32) {
	o.DestinationX = v
}

// GetDestinationY returns the DestinationY field value
func (o *V1Movement) GetDestinationY() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.DestinationY
}

// GetDestinationYOk returns a tuple with the DestinationY field value
// and a boolean to check if the value has been set.
func (o *V1Movement) GetDestinationYOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.DestinationY, true
}

// SetDestinationY sets field value
func (o *V1Movement) SetDestinationY(v int32) {
	o.DestinationY = v
}

// GetDepartureEpoch returns the DepartureEpoch field value
func (o *V1Movement) GetDepartureEpoch() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.DepartureEpoch
}

// GetDepartureEpochOk returns a tuple with the DepartureEpoch field value
// and a boolean to check if the value has been set.
func (o *V1Movement) GetDepartureEpochOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.DepartureEpoch, true
}

// SetDepartureEpoch sets field value
func (o *V1Movement) SetDepartureEpoch(v int64) {
	o.DepartureEpoch = v
}

// GetSpeed returns the Speed field value
func (o *V1Movement) GetSpeed() float64 {
	if o == nil {
		var ret float64
		return ret
	}

	return o.Speed
}

// GetSpeedOk returns a tuple with the Speed field value
// and a boolean to check if the value has been set.
func (o *V1Movement) GetSpeedOk() (*float64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Speed, true
}

// SetSpeed sets field value
func (o *V1Movement) SetSpeed(v float64) {
	o.Speed = v
}

// GetUnitCount returns the UnitCount field value
func (o *V1Movement) GetUnitCount() map[string]int64 {
	if o == nil {
		var ret map[string]int64
		return ret
	}

	return o.UnitCount
}

// GetUnitCountOk returns a tuple with the UnitCount field value
// and a boolean to check if the value has been set.
func (o *V1Movement) GetUnitCountOk() (*map[string]int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.UnitCount, true
}

// SetUnitCount sets field value
func (o *V1Movement) SetUnitCount(v map[string]int64) {
	o.UnitCount = v
}

// GetResourceCount returns the ResourceCount field value
func (o *V1Movement) GetResourceCount() map[string]int64 {
	if o == nil {
		var ret map[string]int64
		return ret
	}

	return o.ResourceCount
}

// GetResourceCountOk returns a tuple with the ResourceCount field value
// and a boolean to check if the value has been set.
func (o *V1Movement) GetResourceCountOk() (*map[string]int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ResourceCount, true
}

// SetResourceCount sets field value
func (o *V1Movement) SetResourceCount(v map[string]int64) {
	o.ResourceCount = v
}

func (o V1Movement) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1Movement) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["id"] = o.Id
	toSerialize["playerID"] = o.PlayerID
	toSerialize["originID"] = o.OriginID
	toSerialize["destinationID"] = o.DestinationID
	toSerialize["destinationX"] = o.DestinationX
	toSerialize["destinationY"] = o.DestinationY
	toSerialize["departureEpoch"] = o.DepartureEpoch
	toSerialize["speed"] = o.Speed
	toSerialize["unitCount"] = o.UnitCount
	toSerialize["resourceCount"] = o.ResourceCount
	return toSerialize, nil
}

func (o *V1Movement) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"playerID",
		"originID",
		"destinationID",
		"destinationX",
		"destinationY",
		"departureEpoch",
		"speed",
		"unitCount",
		"resourceCount",
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

	varV1Movement := _V1Movement{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varV1Movement)

	if err != nil {
		return err
	}

	*o = V1Movement(varV1Movement)

	return err
}

type NullableV1Movement struct {
	value *V1Movement
	isSet bool
}

func (v NullableV1Movement) Get() *V1Movement {
	return v.value
}

func (v *NullableV1Movement) Set(val *V1Movement) {
	v.value = val
	v.isSet = true
}

func (v NullableV1Movement) IsSet() bool {
	return v.isSet
}

func (v *NullableV1Movement) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1Movement(val *V1Movement) *NullableV1Movement {
	return &NullableV1Movement{value: val, isSet: true}
}

func (v NullableV1Movement) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1Movement) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


