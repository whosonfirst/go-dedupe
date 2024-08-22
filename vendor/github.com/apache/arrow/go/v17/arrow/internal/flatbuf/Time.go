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

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

// / Time is either a 32-bit or 64-bit signed integer type representing an
// / elapsed time since midnight, stored in either of four units: seconds,
// / milliseconds, microseconds or nanoseconds.
// /
// / The integer `bitWidth` depends on the `unit` and must be one of the following:
// / * SECOND and MILLISECOND: 32 bits
// / * MICROSECOND and NANOSECOND: 64 bits
// /
// / The allowed values are between 0 (inclusive) and 86400 (=24*60*60) seconds
// / (exclusive), adjusted for the time unit (for example, up to 86400000
// / exclusive for the MILLISECOND unit).
// / This definition doesn't allow for leap seconds. Time values from
// / measurements with leap seconds will need to be corrected when ingesting
// / into Arrow (for example by replacing the value 86400 with 86399).
type Time struct {
	_tab flatbuffers.Table
}

func GetRootAsTime(buf []byte, offset flatbuffers.UOffsetT) *Time {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Time{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Time) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Time) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Time) Unit() TimeUnit {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return TimeUnit(rcv._tab.GetInt16(o + rcv._tab.Pos))
	}
	return 1
}

func (rcv *Time) MutateUnit(n TimeUnit) bool {
	return rcv._tab.MutateInt16Slot(4, int16(n))
}

func (rcv *Time) BitWidth() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 32
}

func (rcv *Time) MutateBitWidth(n int32) bool {
	return rcv._tab.MutateInt32Slot(6, n)
}

func TimeStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func TimeAddUnit(builder *flatbuffers.Builder, unit TimeUnit) {
	builder.PrependInt16Slot(0, int16(unit), 1)
}
func TimeAddBitWidth(builder *flatbuffers.Builder, bitWidth int32) {
	builder.PrependInt32Slot(1, bitWidth, 32)
}
func TimeEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
