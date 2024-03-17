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

// checks if the V1City type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &V1City{}

// V1City struct for V1City
type V1City struct {
	CityInfo V1CityInfo `json:"cityInfo"`
	CityBuildings V1CityBuildings `json:"cityBuildings"`
	CityResources V1CityResources `json:"cityResources"`
	UnitCount map[string]int64 `json:"unitCount"`
}

type _V1City V1City

// NewV1City instantiates a new V1City object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1City(cityInfo V1CityInfo, cityBuildings V1CityBuildings, cityResources V1CityResources, unitCount map[string]int64) *V1City {
	this := V1City{}
	this.CityInfo = cityInfo
	this.CityBuildings = cityBuildings
	this.CityResources = cityResources
	this.UnitCount = unitCount
	return &this
}

// NewV1CityWithDefaults instantiates a new V1City object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1CityWithDefaults() *V1City {
	this := V1City{}
	return &this
}

// GetCityInfo returns the CityInfo field value
func (o *V1City) GetCityInfo() V1CityInfo {
	if o == nil {
		var ret V1CityInfo
		return ret
	}

	return o.CityInfo
}

// GetCityInfoOk returns a tuple with the CityInfo field value
// and a boolean to check if the value has been set.
func (o *V1City) GetCityInfoOk() (*V1CityInfo, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CityInfo, true
}

// SetCityInfo sets field value
func (o *V1City) SetCityInfo(v V1CityInfo) {
	o.CityInfo = v
}

// GetCityBuildings returns the CityBuildings field value
func (o *V1City) GetCityBuildings() V1CityBuildings {
	if o == nil {
		var ret V1CityBuildings
		return ret
	}

	return o.CityBuildings
}

// GetCityBuildingsOk returns a tuple with the CityBuildings field value
// and a boolean to check if the value has been set.
func (o *V1City) GetCityBuildingsOk() (*V1CityBuildings, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CityBuildings, true
}

// SetCityBuildings sets field value
func (o *V1City) SetCityBuildings(v V1CityBuildings) {
	o.CityBuildings = v
}

// GetCityResources returns the CityResources field value
func (o *V1City) GetCityResources() V1CityResources {
	if o == nil {
		var ret V1CityResources
		return ret
	}

	return o.CityResources
}

// GetCityResourcesOk returns a tuple with the CityResources field value
// and a boolean to check if the value has been set.
func (o *V1City) GetCityResourcesOk() (*V1CityResources, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CityResources, true
}

// SetCityResources sets field value
func (o *V1City) SetCityResources(v V1CityResources) {
	o.CityResources = v
}

// GetUnitCount returns the UnitCount field value
func (o *V1City) GetUnitCount() map[string]int64 {
	if o == nil {
		var ret map[string]int64
		return ret
	}

	return o.UnitCount
}

// GetUnitCountOk returns a tuple with the UnitCount field value
// and a boolean to check if the value has been set.
func (o *V1City) GetUnitCountOk() (*map[string]int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.UnitCount, true
}

// SetUnitCount sets field value
func (o *V1City) SetUnitCount(v map[string]int64) {
	o.UnitCount = v
}

func (o V1City) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o V1City) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["cityInfo"] = o.CityInfo
	toSerialize["cityBuildings"] = o.CityBuildings
	toSerialize["cityResources"] = o.CityResources
	toSerialize["unitCount"] = o.UnitCount
	return toSerialize, nil
}

func (o *V1City) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"cityInfo",
		"cityBuildings",
		"cityResources",
		"unitCount",
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

	varV1City := _V1City{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varV1City)

	if err != nil {
		return err
	}

	*o = V1City(varV1City)

	return err
}

type NullableV1City struct {
	value *V1City
	isSet bool
}

func (v NullableV1City) Get() *V1City {
	return v.value
}

func (v *NullableV1City) Set(val *V1City) {
	v.value = val
	v.isSet = true
}

func (v NullableV1City) IsSet() bool {
	return v.isSet
}

func (v *NullableV1City) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1City(val *V1City) *NullableV1City {
	return &NullableV1City{value: val, isSet: true}
}

func (v NullableV1City) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1City) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


