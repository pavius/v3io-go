using Go = import "/go.capnp";
@0xdfe00955984fcb17;

# Imports & Namespace settings
using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("VnObjectItemsGetResponse");

using import "/node/common/VnObjectItemsScanCookie.capnp".VnObjectItemsScanCookie;
using import "/node/common/ExtAttrValue.capnp".ExtAttrValue;
using import "/node/common/VnObjectAttributeMap.capnp".VnObjectAttributeMap;

struct VnObjectItemsGetResponseHeader{
    marker			@0 : Text;
    scanState		@1 : VnObjectItemsScanCookie;
    hasMore         @2 : Bool;
}

struct VnObjectItemsGetMappedKeyValuePair {
	value @0       :ExtAttrValue;
	keyMapIndex @1 :UInt64;
}

struct VnObjectItemsGetItem{
    name        @0 :Text;
    values      @1 :List(VnObjectItemsGetMappedKeyValuePair);
}

# Wrapper so that we can create orphan VnObjectItemsGetItem objects and then fill out a list of pointers
# to them. See https://capnproto.org/faq.html under "How do I resize a list?" (28/08/2016):
# "Keep in mind that you can use orphans to allocate sub-objects before you have a place to put them. But, also
# note that you cannot allocate elements of a struct list as orphans and then put them together as a list later,
# because struct lists are encoded as a flat array of struct values, not an array of pointers to struct values.
# You can, however, allocate any inner objects embedded within those structs as orphans."

struct VnObjectItemsGetItemPtr{
    val @0: VnObjectItemsGetItem;
}


struct VnObjectItemsGetResponsePayload{
	values      @0 :List(VnObjectItemsGetItemPtr);
	keyMap      @1 :VnObjectAttributeMap;
}
