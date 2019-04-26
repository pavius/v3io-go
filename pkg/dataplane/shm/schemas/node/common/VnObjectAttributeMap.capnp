using Go = import "/go.capnp";
@0x986bf57944c8b89f;

using Java = import "/java/java.capnp";
using import "/node/common/StringWrapper.capnp".StringWrapperList;
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("VnObjectAttributeMapOuter");

struct VnObjectAttributeMap {
	names @0 : StringWrapperList;
}
