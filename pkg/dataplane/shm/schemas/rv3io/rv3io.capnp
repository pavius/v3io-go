using Go = import "/go.capnp";
@0x8ae7c3caa64d402e;

# Go stuff
$Go.package("rv3io_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/rv3io");

# Imports & Namespace settings
using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("Rv3ioCommands");

using import "/node/common/FileSystemAttrGetRequest.capnp".FileSystemAttrGetRequest;
using import "/node/common/FileSystemAttrGetResponse.capnp".FileSystemAttrGetResponse;
using import "/node/common/FileSystemSetAttrRequest.capnp".FileSystemSetAttrRequest;
using import "/node/common/TinyData.capnp".TinyData;
using import "/node/common/TimeSpec.capnp".TimeSpec;
using import "/node/common/StringWrapper.capnp".StringWrapper;
using import "/node/common/InterfaceType.capnp".InterfaceType;
using import "/node/common/ExtAttr.capnp".ExtAttrs;
using import "/node/common/ExtAttrKey.capnp".ExtAttrKey;

const rv3ioVersionMajor :UInt32 = 1;
const rv3ioVersionMinor :UInt32 = 28;

struct Rv3ioHelloRequest{
    writeShmemBytes             @0 : UInt64=0;
    protocolVersionMajor        @1 : UInt32 = .rv3ioVersionMajor;
    protocolVersionMinor        @2 : UInt32 = .rv3ioVersionMinor;
}

struct Rv3ioSessionAcquireRequest {
    interfaceType               @0 : InterfaceType;
    userName                    @1 : Text = "";
    password                    @3 : Text = "";
    label                       @2 : Text = "";
}

struct Rv3ioChannelCreateRequest {
    workerPreference            @0 : UInt64;
    consumerNumQuervoItems      @1 : UInt64 = 8192;
    producerNumQuervoItems      @2 : UInt64 = 8192;
}

struct Rv3ioChannelCreateResponse {
    channelId                   @0 : UInt64;
    workerId                    @1 : UInt64;
    consumerShmPath             @2 : Text = "";
    producerShmPath             @3 : Text = "";
    heapShmPath                 @4 : Text = "";
    heapShmPaths                @5 : List(Text);
}

struct Rv3ioChannelDeleteRequest {
    channelId                   @0 : UInt64;
}

struct Rv3ioContainerOpenRequest{
    containerId                 @0 : Text = "";
    containerAlias              @1 : Text = "";
}

struct Rv3ioContainerOpenResponse{
    containerHandle             @0 : UInt64 = 0;
}

struct RV3ioSharedBuffer{
    sharedFilePath              @0 : Text = "";
    offset                      @1 : UInt64 = 0;
    length                      @2 : UInt64 = 0;
    sharedFileId                @3 : UInt64 = 0;
}

struct Rv3ioHelloResponse{
    readBuffer                  @0 : RV3ioSharedBuffer;
    writeBuffer                 @1 : RV3ioSharedBuffer;
}

struct Rv3ioSessionAcquireResponse{
    sessionId                   @0 : Rv3ioAuthSession;
}

struct Rv3ioFileOpenRequest{
    oflags                      @0 : UInt64 = 0;
    mode                        @1 : UInt64 = 0;
    path                        @2 : Text = "";
}

struct Rv3ioFileOpenResponse{
    fileHandle                  @0 : UInt64 = 0;
}

struct Rv3ioFileCloseRequest{
    fileHandle                  @0 : UInt64 = 0;
}

struct Rv3ioFileReadRequest{
    fileHandle                  @0 : UInt64 = 0;
    offset                      @1 : UInt64 = 0;
    bytesCount                  @2 : UInt64 = 0;
}

struct Rv3ioFileReadResponse{
    bytesRead                   @0 : UInt64 = 0;
    offsetInSharedFile          @1 : UInt64 = 0;
    memHandle                   @2 : UInt64 = 0;
}

struct Rv3ioFileWriteRequest{
    fileHandle                  @0 : UInt64 = 0;
    offset                      @1 : UInt64 = 0;
    offsetInSharedFile          @2 : UInt64 = 0;
    bytesCount                  @3 : UInt64 = 0;
    sharedFileId                @4 : UInt64 = 0;
}

struct Rv3ioFileWriteResponse{
    bytesWritten                @0 : UInt64 = 0;
}

enum Rv3ioDirReadType {
    allKeys             @0;
    dirOnly             @1;
}

struct Rv3ioDirReadRequest{
    maxObjects                  @0 : UInt64 = 0;
    path                        @1 : Text = "";
    marker                      @2 : Text = "";
    type                        @3 : Rv3ioDirReadType;
}

struct Rv3ioObjectEntry{
    name                        @0 : Text = "";
    size                        @1 : UInt64 = 0;
    mode                        @2 : UInt64 = 0;
    atime                       @3 : UInt64 = 0;
    ctime                       @4 : UInt64 = 0;
    mtime                       @5 : UInt64 = 0;
    streamSeqId                 @6 : UInt64 = 0;
    gid                         @7 : UInt32 = 0;
    uid                         @8 : UInt32 = 0;
    inodeNumber                 @9: UInt64 = 0;
}

struct Rv3ioDirReadResponseHeader{
    bytesRead                   @0 : UInt64 = 0;
    offsetInSharedFile          @1 : UInt64 = 0;
    memHandle                   @2 : UInt64 = 0;
    remainingObjects            @3 : UInt64 = 0;
    numItems                    @4 : UInt64 = 0;
}

struct Rv3ioDirReadResponseBody{
    objects                     @0 : List(Rv3ioObjectEntry);
}

struct Rv3ioDirCreateRequest{
    mode                        @0 : UInt64 = 0;
    path                        @1 : Text = "";
}

struct Rv3ioDirDeleteRequest{
    path                        @0 : Text = "";
}

struct Rv3ioFsFileDeleteRequest{
    path                        @0 : Text = "";
}

struct Rv3ioRenameRequest{
    oldPath                     @0 : Text = "";
    newPath                     @1 : Text = "";
}

struct Rv3ioStreamCreateRequest {
    mode                        @0 : UInt64 = 0;
    path                        @1 : Text = "";
    shardCount                  @2 : UInt16 = 0;
    union {
        shardPeriodSec          @3 : UInt32 = 0;
        shardPeriodHour         @4 : UInt32 = 0;
    }
}

struct Rv3ioObjectPutRequest {
    name                        @0 : Text = "";
    bytesToWrite                @1 : UInt64 = 0;
    offset                      @2 : Int64 = 0;
    offsetInSharedFile          @3 : UInt64 = 0;
}

struct Rv3ioObjectPutResponse{
    bytesWritten    @0 :UInt64;
}


struct Rv3ioObjectGetRequest {
    name            @0 :Text;
    bytesNum        @1 :UInt64;
    offset          @2 :UInt64;
}

struct Rv3ioObjectGetResponse {
    bytesRead               @0 : UInt64 = 0;
    bytesRemain             @1 : UInt64 = 0;
    offsetInSharedFile      @2 : UInt64 = 0;
    memHandle               @3 : UInt64 = 0;
}

enum Rv3ioValidationOperationType {
    validateEq          @0;
    validateNeq         @1;
    validateLtOrEq      @2;
    validateLt          @3;
    validateGtOrEq      @4;
    validateGt          @5;
}

enum Rv3ioSetOperationType {
    operationUpdate      @0;
    operationReset       @1;
    operationInc         @2;
    operationDec         @3;
}

struct Rv3ioObjectSetRequest {
    name                        @0 : Text;
    validationMTime             @1 : TimeSpec;
    validationOperationType     @2 : Rv3ioValidationOperationType;
    validationMask              @3 : TinyData;
    validationValue             @4 : TinyData;
    setOperationType            @5 : Rv3ioSetOperationType;
    dataMask                    @6 : TinyData;
    dataValue                   @7 : TinyData;
}

struct Rv3ioObjectSetResponse{
    objectSet           @0 : Bool;
    preSetData          @1 : TinyData;
    postSetData         @2 : TinyData;
    mtime               @3 : TimeSpec;
    prevMtime           @4 : TimeSpec;
}

struct Rv3ioRecordsGetRequest {
    path                @0 : Text = "";
    shardLocation       @1 : Data;
    bytesNum            @2 : UInt64 = 0;
    startSeqId          @3 : UInt64 = 0;
    maxNumRecords       @4 : UInt32 = 0;
}

struct Rv3ioRecordsGetResponse {
    nextShardLocation   @0 : Data;
    numRecordsReturn    @1 : UInt64 = 0;
    lagMillisecond      @2 : UInt64 = 0;
    lagRecords          @3 : UInt64 = 0;
    sharedMemDetails    @4 : Rv3ioFileReadResponse; #no need for this struct once moved to new daemon
    bytesRead           @5 : UInt64 = 0; #same as bytesRead in Rv3ioFileReadResponse
}

enum Rv3ioSeekType {
        invalid         @0;
        sequenceBase    @1;
        latest          @2;
        earliest        @3;
        time            @4;
}

struct Rv3ioShardSeekRequest {
    path                @0 : Text = "";
    seekType            @1 : Rv3ioSeekType;
    union {
        none            @2 : Void;
        startSeqId      @3 : UInt64 = 0;
        startTime       @4 : TimeSpec;
    }
}

struct Rv3ioShardSeekResponse {
    shardLocation          @0 : Data;
}

#---------------------------- Put Records - Begin ------------------------------
struct Rv3ioObjectOnWriteSharedFile {
    offsetInSharedFile  @0 : UInt64 = 0;
    length              @1 : UInt64 = 0;
}

struct Rv3ioPutRecordHeaderList {
    headers     @0 : List(Rv3ioPutRecordHeader);
}

struct Rv3ioPutRecordResponseList {
    headers     @0 : List(Rv3ioRecordPutResult);
}

struct Rv3ioPutRecordHeader {
    # should not be passed over network. Use this struct for serialization only.
    sequenceIdInBatch   @0 : UInt32 = 0;
    payloadSize         @1 : UInt32 = 0;
    payloadOffset       @2 : UInt64 = 0; # the absolute offset on shared memory
    partitionId         @3 : Int16 = -1; # Set invalid shard id by default
    key                 @4 : Text = "";
    clientInfo          @5 : Data = "";
}

struct Rv3ioRecordsPutRequest {
    topicPath                   @0 : Text;
    headersOffsetInSharedFile   @1 : Rv3ioObjectOnWriteSharedFile; # the absolute offset of struct Rv3ioPutRecordHeaderList
    totalPayloadSize            @2 : UInt64 = 0; # the total size of payloads (including the padding) in Bytes
    totalHeaderListElem         @3 : UInt32 = 0; # the total amount of headers in the list
}

struct RecordPutSuccessResult {
    sequenceId          @0 : UInt64 = 0; # sequence ID of the record in the stream
    partitionId         @1 : UInt16 = 0; # shard ID
}

struct RecordPutFailureResult {
    errorCode           @0 : Int64 = -1;
}

struct Rv3ioRecordPutResult {
    sequenceIdInBatch   @0 : UInt32 = 0; # unique identifier of the record within the batch
    union{
      success           @1 : RecordPutSuccessResult;
      failure           @2 : RecordPutFailureResult;
    }
}

struct Rv3ioRecordsPutResponse {
    # offsetInSharedFile is pointing on the beginning of the list of Rv3ioRecordPutResult
    offsetInSharedFile  @0 : UInt64 = 0;
    memHandle           @1 : UInt64 = 0; # memory handler to release
    bytesRead           @2 : UInt64 = 0;
    numRecords          @3 : UInt32 = 0;
}
#----------------------------- Put Records - End -------------------------------

struct Rv3ioObjectItemOnWriteSharedFile {
    offsetInSharedFile  @0 : UInt64 = 0;
    length              @1 : UInt64 = 0;
}

struct Rv3ioObjectItemOnReadSharedFile {
    offsetInSharedFile          @0 : UInt64 = 0;
    length                      @1 : UInt64 = 0;
    memHandle                   @2 : UInt64 = 0;
}

struct Rv3ioObjectItemsGetCookie {
    data    @0 : Data; # according to v3io_parallel_scan_execution_cookie - Text length must be as the cookie in v3io
}

struct Rv3ioObjectItemParallelExecution {
    partNum     @0 : UInt16 = 0;
    partIndex   @1 : UInt16 = 0;
    union {
        none            @2 : Void;
        partCookie      @3 : Rv3ioObjectItemsGetCookie;
    }
}

struct Rv3ioObjectItemsGetRequest {
    name                        @0 : Text;
    startAfterKey               @1 : Text;
    endKey                      @2 : Text;
    maxObjects                  @3 : UInt64 = 0;
    payloadBufferSize           @4 : UInt64 = 0; # not used in new daemon
    parallelExecution           @5 : Rv3ioObjectItemParallelExecution;
    attributes : union {
        attributesKeysOffset   @6 : Rv3ioObjectItemOnWriteSharedFile; # the serialized struct CollectionAttributeKeysList
        allKeys                @9 : Void;
        allKeysWithSpecial     @10: Void;
        attributesList         @11: Void; # attributes are in payload. for the new daemon
    }
    filterOffset                @7 : Rv3ioObjectItemOnWriteSharedFile; # the serialized CollectionFilter struct
    sharedFileId                @8 : UInt64 = 0;
}

struct Rv3ioObjectItemsGetRequestPayload {
    attributesKeys         @0: Data; # the serialized struct CollectionAttributeKeysList for new daemon
    filter                 @1: Data; # the serialized CollectionFilter struct, for the new daemon
}

struct CollectionAttributeKeysList {
    attributes          @0 : List(StringWrapper);
}

struct Rv3ioObjectItemsGetResponse {
    startScanAfter      @0 : Text;
    partCookie          @1 : Rv3ioObjectItemsGetCookie;
    hasMore             @2 : Bool = false;
    payloadOffset       @3 : Rv3ioObjectItemOnReadSharedFile; # serialized VnObjectItemsGetResponsePayload struct
}

enum Rv3ioObjectItemUpdateMode {
    overwriteRow                @0;
    appendAttributesOrCreateRow @1; # if row exist - append only new fields, else - put new record
    appendOnlyNewRow            @2; # if row exist - do nothing, else - put new record
    appendOrReplaceAttributes   @3; # Either row exist or not - append new fields and replace existing
}

struct Rv3ioObjectItemUpdateRequest {
    name                        @0 : Text;
    key                         @1 : Text;
    mode                        @2 : Rv3ioObjectItemUpdateMode;
    requestPayloadOffset        @3 : Rv3ioObjectItemOnWriteSharedFile; # the VnObjectItemUpdateRequestPayload 
    sharedFileId                @4 : UInt64 = 0;
}

struct Rv3ioObjectItemUpdateResponse {
    itemHasBeenUpdated          @0 : Bool;
}

struct Rv3ioObjectItemBatchUpdateRequest {
    # the serialized List(Rv3ioObjectItemOnSharedFile) - each entry in the list is an offset to Rv3ioObjectItemUpdateRequest
    updateRequestListOffset     @0 : Rv3ioObjectItemOnReadSharedFile;
}

struct Rv3ioObjectItemBatchUpdateEntryResponse {
    requestOffset   @0 : UInt64 = 0;
    retCode             @1 : Int64 = 0;
}

struct Rv3ioObjectItemBatchUpdateResponse {
    resultListOffset    @0 : Rv3ioObjectItemOnReadSharedFile; # the serialized List(Rv3ioObjectItemBatchUpdateEntryResponse);
    memHandle           @1 : UInt64 = 0;
}

struct Rv3ioObjectDeleteRequest {
    name               @0 :Text;
}

struct Rv3ioAuthorizationPosix {
    gid                @0 : UInt32 = 0;
    uid                @1 : UInt32 = 0;
}

struct Rv3ioAuthSession{
    sessionId          @0 : UInt32 = 0;
}

struct Rv3ioAuthorization {
    union {
       none            @0 : Void;
       posix           @1 : Rv3ioAuthorizationPosix;
       session         @2 : Rv3ioAuthSession;
    }
}

struct Rv3ioObjectItemGetRequest {
    objectName          @0 : Text;
    payloadSize         @1 : UInt64 = 0;
    union {
        keys                @2 : List(ExtAttrKey);
        allKeys             @3 : Void;
        allKeysWithSpecial  @4 : Void;
    }
}

struct Rv3ioObjectItemGetResponse {
    offsetInSharedFile          @0 : UInt64 = 0; # points to serialized ExtAttrs list
    length                      @1 : UInt64 = 0;
    memHandle                   @2 : UInt64 = 0;
}


struct Rv3ioRequest{

    containerHandle             @0 : UInt64; # not for hello/container open
    handlesToFree               @1 : List(UInt64);
    transactionId               @2 : UInt64 = 0;
    authorization               @3 : Rv3ioAuthorization;

    union{
        hello                   @4 : Rv3ioHelloRequest;
        fileOpen                @5 : Rv3ioFileOpenRequest;
        fileClose               @6 : Rv3ioFileCloseRequest;
        fileRead                @7 : Rv3ioFileReadRequest;
        fileWrite               @8 : Rv3ioFileWriteRequest;
        containerOpen           @9 : Rv3ioContainerOpenRequest;
        containerClose          @10 : Void;
        fsAttrGet               @11 : FileSystemAttrGetRequest;
        fsDirRead               @12 : Rv3ioDirReadRequest;
        fsDirCreate             @13 : Rv3ioDirCreateRequest;
        fsDirDelete             @14 : Rv3ioDirDeleteRequest;
        rename                  @15 : Rv3ioRenameRequest;
        fsFileDelete            @16 : Rv3ioFsFileDeleteRequest;
        objectPut               @17 : Rv3ioObjectPutRequest;
        objectGet               @18 : Rv3ioObjectGetRequest;
        objectDelete            @19 : Rv3ioObjectDeleteRequest;
        objectSet               @20 : Rv3ioObjectSetRequest;
        recordsGet              @21 : Rv3ioRecordsGetRequest;
        none                    @22 : Void;
        streamCreate            @23 : Rv3ioStreamCreateRequest;
        streamSeekShard         @24 : Rv3ioShardSeekRequest;
        recordsPut              @25 : Rv3ioRecordsPutRequest;
        objectItemsGet          @26 : Rv3ioObjectItemsGetRequest;
        objectItemUpdate        @27 : Rv3ioObjectItemUpdateRequest;
        fsAttrSet               @28 : FileSystemSetAttrRequest;
        sessionAcquire          @29 : Rv3ioSessionAcquireRequest;
        objectItemGet           @30 : Rv3ioObjectItemGetRequest;
        channelCreate           @31 : Rv3ioChannelCreateRequest;
        channelDelete           @32 : Rv3ioChannelDeleteRequest;
        last                    @33 : Void; # always make sure last's capnp tag (@XX number) is highest in the union
    }
}

struct Rv3ioResponse{

    result                      @0 :Int64 = -1;
    freeFails                   @1 :List(UInt64);
    errMsg                      @2 :Text = "";
    transactionId               @3: UInt64 = 0;

    union{
        nothing                 @4 : Void;
        hello                   @5 : Rv3ioHelloResponse;
        fileOpen                @6 : Rv3ioFileOpenResponse;
        fileRead                @7 : Rv3ioFileReadResponse;
        fileWrite               @8 : Rv3ioFileWriteResponse;
        containerOpen           @9 : Rv3ioContainerOpenResponse;
        fsAttrGet               @10 : FileSystemAttrGetResponse;
        fsReadDir               @11 : Rv3ioDirReadResponseHeader;
        objectPut               @12 : Rv3ioObjectPutResponse;
        objectGet               @13 : Rv3ioObjectGetResponse;
        objectSet               @14 : Rv3ioObjectSetResponse;
        recordsGet              @15 : Rv3ioRecordsGetResponse;
        streamSeekShard         @16 : Rv3ioShardSeekResponse;
        recordsPut              @17 : Rv3ioRecordsPutResponse;
        objectItemsGet          @18 : Rv3ioObjectItemsGetResponse;
        objectItemUpdate        @19 : Rv3ioObjectItemUpdateResponse;
        sessionAcquire          @20 : Rv3ioSessionAcquireResponse;
        objectItemGet           @21 : Rv3ioObjectItemGetResponse;
        channelCreate           @22 : Rv3ioChannelCreateResponse;
        last                    @23 : Void; # always make sure last's capn tag (@XX number) is highest in the union
    }
}