using Go = import "/go.capnp";
@0x8e1d61a4db048b89;

using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("VnObjectItemUpdateRequestOuter");

using import "/node/common/CollectionFilter.capnp".CollectionFilter;
using import "/node/common/ExtAttr.capnp".ExtAttrs;
using import "/node/common/TinyData.capnp".TinyData;
using import "/node/common/ShardData.capnp".ShardData;

enum CollectionItemUpdateMode {
	createRowIfNotExists 	@0;
	overwriteEntireRow 	@1;
	appendNonExistingColumns 	@2;
    appendOrReplaceColumns 	@3;
}

enum InodeType {
    inodeTypeRegular        @0;
    inodeTypeTiny           @1;
    inodeTypeStream         @2;
}

struct VnObjectItemUpdateRequest{
	collectionName  @0 :Text;
	itemKey   	@1 :Text;
	mode   		@2 :CollectionItemUpdateMode;
	inodeType   @3 :InodeType;
	permissionMode @4 :UInt32;
	inodeSizeOnCreate @5 : UInt64;
	union {
		valueTinyOnCreate       @6 : TinyData;
		shardDataCreate         @7 :ShardData;
	}
}

struct VnObjectItemUpdateRequestPayload{
        union {
		attributes          @0 :ExtAttrs;
		updateText          @1 :Text;
	}

	filter              @2 :CollectionFilter;
}
