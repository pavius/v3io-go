/*
Copyright 2018 The v3io Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v3io

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"os"
	"time"
)

//
// Control plane
//

type NewContextInput struct {
	ClusterEndpoints []string
	NumWorkers       int
	RequestChanLen   int
	TlsConfig        *tls.Config
	DialTimeout      time.Duration
}

type NewSessionInput struct {
	Username  string
	Password  string
	AccessKey string
}

type NewContainerInput struct {
	ContainerName string
}

//
// Data plane
//

type DataPlaneInput struct {
	Ctx                 context.Context
	ContainerName       string
	AuthenticationToken string
	AccessKey           string
	Timeout             time.Duration
}

type DataPlaneOutput struct {
	ctx context.Context
}

// TODO: probably make this transport specific
type Statistics struct {
	NumCommandsProduced     uint64
	NumCommandsConsumed     uint64
	NumAllocationRequests   uint64
	NumAllocationResponses  uint64
	NumJobRequests          uint64
	NumJobResponses         uint64
	NumDeallocationRequests uint64
}

//
// Container
//

type GetContainerContentsInput struct {
	DataPlaneInput
	Path string
}

type Content struct {
	XMLName        xml.Name `xml:"Contents"`
	Key            string   `xml:"Key"`
	Size           int      `xml:"Size"`
	LastSequenceID int      `xml:"LastSequenceId"`
	ETag           string   `xml:"ETag"`
	LastModified   string   `xml:"LastModified"`
}

type CommonPrefix struct {
	CommonPrefixes xml.Name `xml:"CommonPrefixes"`
	Prefix         string   `xml:"Prefix"`
}

type GetContainerContentsOutput struct {
	BucketName     xml.Name       `xml:"ListBucketResult"`
	Name           string         `xml:"Name"`
	NextMarker     string         `xml:"NextMarker"`
	MaxKeys        string         `xml:"MaxKeys"`
	Contents       []Content      `xml:"Contents"`
	CommonPrefixes []CommonPrefix `xml:"CommonPrefixes"`
}

type GetContainersInput struct {
	DataPlaneInput
}

type GetContainersOutput struct {
	DataPlaneOutput
	XMLName xml.Name    `xml:"ListAllMyBucketsResult"`
	Owner   interface{} `xml:"Owner"`
	Results Containers  `xml:"Buckets"`
}

type Containers struct {
	Name       xml.Name        `xml:"Buckets"`
	Containers []ContainerInfo `xml:"Bucket"`
}

type ContainerInfo struct {
	BucketName   xml.Name `xml:"Bucket"`
	Name         string   `xml:"Name"`
	CreationDate string   `xml:"CreationDate"`
	ID           int      `xml:"Id"`
}

//
// Object
//

type GetObjectInput struct {
	DataPlaneInput
	Path     string
	Offset   int
	NumBytes int
}

type PutObjectInput struct {
	DataPlaneInput
	Path   string
	Offset int
	Body   []byte
}

type DeleteObjectInput struct {
	DataPlaneInput
	Path string
}

//
// KV
//

type PutItemInput struct {
	DataPlaneInput
	Path       string
	Condition  string
	Attributes map[string]interface{}
}

type PutItemsInput struct {
	DataPlaneInput
	Path      string
	Condition string
	Items     map[string]map[string]interface{}
}

type PutItemsOutput struct {
	DataPlaneOutput
	Success bool
	Errors  map[string]error
}

type UpdateItemInput struct {
	DataPlaneInput
	Path       string
	Attributes map[string]interface{}
	Expression *string
	Condition  string
}

type GetItemInput struct {
	DataPlaneInput
	Path           string
	AttributeNames []string
}

type GetItemOutput struct {
	DataPlaneOutput
	Item Item
}

type GetItemsInput struct {
	DataPlaneInput
	Path              string
	AttributeNames    []string
	Filter            string
	Marker            string
	ShardingKey       string
	Limit             int
	Segment           int
	TotalSegments     int
	SortKeyRangeStart string
	SortKeyRangeEnd   string
}

type GetItemsOutput struct {
	DataPlaneOutput
	Last       bool
	NextMarker string
	Items      []Item
}

//
// Stream
//

type StreamRecord struct {
	ShardID      *int
	Data         []byte
	ClientInfo   []byte
	PartitionKey string
}

type SeekShardInputType int

const (
	SeekShardInputTypeTime SeekShardInputType = iota
	SeekShardInputTypeSequence
	SeekShardInputTypeLatest
	SeekShardInputTypeEarliest
)

type CreateStreamInput struct {
	DataPlaneInput
	Path                 string
	ShardCount           int
	RetentionPeriodHours int
}

type DeleteStreamInput struct {
	DataPlaneInput
	Path string
}

type PutRecordsInput struct {
	DataPlaneInput
	Path    string
	Records []*StreamRecord
}

type PutRecordResult struct {
	SequenceNumber int
	ShardID        int `json:"ShardId"`
	ErrorCode      int
	ErrorMessage   string
}

type PutRecordsOutput struct {
	DataPlaneOutput
	FailedRecordCount int
	Records           []PutRecordResult
}

type SeekShardInput struct {
	DataPlaneInput
	Path                   string
	Type                   SeekShardInputType
	StartingSequenceNumber int
	Timestamp              int
}

type SeekShardOutput struct {
	DataPlaneOutput
	Location string
}

type GetRecordsInput struct {
	DataPlaneInput
	Path     string
	Location string
	Limit    int
}

type GetRecordsResult struct {
	ArrivalTimeSec  int
	ArrivalTimeNSec int
	SequenceNumber  int
	ClientInfo      []byte
	PartitionKey    string
	Data            []byte
}

type GetRecordsOutput struct {
	DataPlaneOutput
	NextLocation        string
	MSecBehindLatest    int
	RecordsBehindLatest int
	Records             []GetRecordsResult
}

//
// File
//

type FileOpenInput struct {
	DataPlaneInput
	FilePath string
	Flags    int
	Mode     os.FileMode
}

type FileOpenOutput struct {
	DataPlaneOutput
	FileHandle uint64
}

type FileCloseInput struct {
	DataPlaneInput
	FileHandle uint64
}

type FileWriteInput struct {
	DataPlaneInput
	FileHandle uint64
	Offset     uint64
	Data       []byte
	Writer     RequestWriter
}

type FileWriteOutput struct {
	DataPlaneOutput
	BytesWritten uint64
}

type FileReadInput struct {
	DataPlaneInput
	FileHandle uint64
	Offset     uint64
	BytesCount uint64
}

type FileReadOutput struct {
	DataPlaneOutput
	Data []byte
}

//
// Shm specific
//

type EchoInput struct {
	DataPlaneInput
	DataUint64 uint64
	DataBytes  []byte
}

type EchoOutput struct {
	DataPlaneOutput
	DataUint64 uint64
	DataBytes  []byte
}

type SessionAcquireInput struct {
	DataPlaneInput
	Username      string
	Password      string
	Label         string
	InterfaceType uint64
}

type SessionAcquireOutput struct {
	DataPlaneOutput
	SessionID uint32
}

type ContainerOpenInput struct {
	DataPlaneInput
	ContainerID    uint64
	ContainerAlias string
}

type ContainerOpenOutput struct {
	DataPlaneOutput
	ContainerHandle uint64
}
