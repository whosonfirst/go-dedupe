package id

import (
	"github.com/whosonfirst/go-whosonfirst-feature/constants"
)

// MULTIPLE_COUNTIES is the identifier used to indicate that a place is legitimately parented by multiple counties.
// For details consult https://github.com/whosonfirst/py-mapzen-whosonfirst-hierarchy/issues/1
// This constant is DEPRECATED. Please use the equivalent constant in whosonfirst/go-whosonfirst-feature/constants
const MULTIPLE_COUNTIES int64 = constants.MULTIPLE_COUNTIES

// MULTIPLE_PARENTS was a misnamed constant and is included for backward compatibility but should otherwise not be used.
// This constant is DEPRECATED. Please use the equivalent constant in whosonfirst/go-whosonfirst-feature/constants
const MULTIPLE_PARENTS int64 = constants.MULTIPLE_PARENTS

// MULTIPLE_NEIGHBOURHOODS is the identifier used to indicate that a place is legitimately parented by multiple neighbouhoods.
// This constant is DEPRECATED. Please use the equivalent constant in whosonfirst/go-whosonfirst-feature/constants
const MULTIPLE_NEIGHBOURHOODS int64 = constants.MULTIPLE_NEIGHBOURHOODS

// ITS_COMPLICATED is the identifier used to indicate that the parentage of a place is complicated in a geopolitical way too nuanced and complex to express otherwise.
// This constant is DEPRECATED. Please use the equivalent constant in whosonfirst/go-whosonfirst-feature/constants
const ITS_COMPLICATED = constants.ITS_COMPLICATED

// UNKNOWN is the identifier used to indicate that an otherwise valid identifier is unknown and needs to be resolved.
// This constant is DEPRECATED. Please use the equivalent constant in whosonfirst/go-whosonfirst-feature/constants
const UNKNOWN int64 = constants.UNKNOWN

// EARTH is the Who's On First identifier for the planet Earth.
// This constant is DEPRECATED. Please use the equivalent constant in whosonfirst/go-whosonfirst-feature/constants
const EARTH int64 = constants.EARTH

// NULL_ISLAND is the Who's On First identifier for the Null Island.
// This constant is DEPRECATED. Please use the equivalent constant in whosonfirst/go-whosonfirst-feature/constants
const NULL_ISLAND int64 = constants.NULL_ISLAND
