@0xd010a72957cfc2e0;
using import "/node/common/Timestamp.capnp".Timestamp;
using import "/node/common/StringWrapper.capnp".StringWrapperList;

# Go stuff
$Go.package("node_common_capnp");
$Go.import("github.com/v3io/v3io-go/internal/schemas/node/common");

using Java = import "/java/java.capnp";
$Java.package("io.iguaz.v3io.daemon.client.api.capnp");
$Java.outerClassname("CollectionFilters");

using import "/node/common/TimeSpec.capnp".TimeSpec;


struct CollectionFilterAstConst{
	union {
		str		@0: Text;
		lst		@1: StringWrapperList;
		int64val	@2: Int64;
		uint64val	@3: UInt64;
		timestamp	@4: TimeSpec;
		doubleVal	@5: Float64;
		boolean     @6: Bool;

	}
}

enum CollectionFilterAttributeType {
	string		@0;
	int64		@1;
	uint64		@2;
	timestamp	@3;
	double		@4;
	boolean     @5;
}

struct CollectionFilterAstAttribute
{
	name		@0: Text;
	type		@1: CollectionFilterAttributeType;
}


enum CollectionFilterOperator {
	stringContains 		@0;
	stringEndsWith 		@1;
	stringStartsWith 	@2;
	not 			@3;
	or 			@4;
	and 			@5;
	exists 			@6;
	in 			@7;
	lessThanOrEqual 	@8;
	lessThan 		@9;
	greaterThanOrEqual 	@10;
	greaterThan 		@11;
	equalTo 		@12;
}

struct CollectionFilterAstOperator {
	op	@0 : CollectionFilterOperator;
	left	@1 : CollectionFilterAstNode;
	right	@2 : CollectionFilterAstNode;
}

struct CollectionFilterAstNode {
    union {
		oper		@0 :CollectionFilterAstOperator;
		constLeaf	@1 :CollectionFilterAstConst;
		attrLeaf	@2 :CollectionFilterAstAttribute;
		nil		@3 :Void;
	}
}

struct CollectionFilter {
	union {
		sqlText			@0 :Text;
		astRoot			@1 :CollectionFilterAstNode;
	}
}
