// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package flatbuf

import "strconv"

// / Represents Arrow Features that might not have full support
// / within implementations. This is intended to be used in
// / two scenarios:
// /  1.  A mechanism for readers of Arrow Streams
// /      and files to understand that the stream or file makes
// /      use of a feature that isn't supported or unknown to
// /      the implementation (and therefore can meet the Arrow
// /      forward compatibility guarantees).
// /  2.  A means of negotiating between a client and server
// /      what features a stream is allowed to use. The enums
// /      values here are intented to represent higher level
// /      features, additional details maybe negotiated
// /      with key-value pairs specific to the protocol.
// /
// / Enums added to this list should be assigned power-of-two values
// / to facilitate exchanging and comparing bitmaps for supported
// / features.
type Feature int64

const (
	/// Needed to make flatbuffers happy.
	FeatureUNUSED Feature = 0
	/// The stream makes use of multiple full dictionaries with the
	/// same ID and assumes clients implement dictionary replacement
	/// correctly.
	FeatureDICTIONARY_REPLACEMENT Feature = 1
	/// The stream makes use of compressed bodies as described
	/// in Message.fbs.
	FeatureCOMPRESSED_BODY Feature = 2
)

var EnumNamesFeature = map[Feature]string{
	FeatureUNUSED:                 "UNUSED",
	FeatureDICTIONARY_REPLACEMENT: "DICTIONARY_REPLACEMENT",
	FeatureCOMPRESSED_BODY:        "COMPRESSED_BODY",
}

var EnumValuesFeature = map[string]Feature{
	"UNUSED":                 FeatureUNUSED,
	"DICTIONARY_REPLACEMENT": FeatureDICTIONARY_REPLACEMENT,
	"COMPRESSED_BODY":        FeatureCOMPRESSED_BODY,
}

func (v Feature) String() string {
	if s, ok := EnumNamesFeature[v]; ok {
		return s
	}
	return "Feature(" + strconv.FormatInt(int64(v), 10) + ")"
}
