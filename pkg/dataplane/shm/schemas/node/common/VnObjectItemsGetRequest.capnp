using Go = import "/go.capnp";
@0xb5cff4cd5131f6ae;

using import "/node/common/VnObjectItemsScanCookie.capnp".VnObjectItemsScanCookie;
using import "/node/common/VnObjectItemsScanCookie.capnp".VnObjectItemsScanStateCapn;
using import "/node/common/ExtAttrKey.capnp".ExtAttrKey;

struct VnObjectItemsGetRequestScanInfo{
    scanState               @0 :VnObjectItemsScanCookie;
}

struct VnObjectItemsFilter{
    # offset from the start of this message to a buffer containing a full message rooted at an CollectionFilter object
    collectionFilter @0 :Data;
}

struct VnObjectItemsGetRequest{
    collectionName          @0 :Text;
    startKey                @1 :Text;
    endKey                  @2 :Text;
    maxObjects              @3 :UInt32;
    union {
        invalid             @4 :Void;
        scan                @5 :VnObjectItemsGetRequestScanInfo;
        query               @6 :Void;
    }
}

struct VnObjectItemsGetRequestPayload{
    filter                  @0 :VnObjectItemsFilter;
    attributes :union {
            keys                @1 :List(ExtAttrKey);
            allKeys             @2 :Void;
            allKeysWithSpecial  @3 :Void;
    }
}
