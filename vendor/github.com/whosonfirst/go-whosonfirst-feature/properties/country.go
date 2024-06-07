package properties

import (
	"github.com/tidwall/gjson"
)

/*

https://github.com/whosonfirst/whosonfirst-placetypes#iso-country-codes

Per the ISO 3166 spec which states:

User-assigned code elements are codes at the disposal of users who need to add further names of countries, territories, or other geographical entities to their in-house application of ISO 3166-1, and the ISO 3166/MA will never use these codes in the updating process of the standard. The following codes can be user-assigned:[19]

    Alpha-2: AA, QM to QZ, XA to XZ, and ZZ
    Alpha-3: AAA to AAZ, QMA to QZZ, XAA to XZZ, and ZZA to ZZZ
    Numeric: 900 to 999

We use the following ISO country codes:

XK	We just followed Geonames' lead and have assigned XK to be the ISO country code for Kosovo.
XN	For Null Island.
XS	We use XS to indicate Somaliland.
XX	XX denotes a place disputed by two or more (ISO) countries.
XY	XY denotes an ISO country that has yet to be determined (by us). You might typically see this is a record for a freshly created place that hasn't been fully vetted or editorialized yet.
XZ	XZ is the ISO country code equivalent of wof:parent_id=-2 or :shrug: the world is a complicated place.

*/

const COUNTRY_KOSOVO string = "XK"
const COUNTRY_NULLISLAND string = "XN"
const COUNTRY_SOMALILAND string = "XS"
const COUNTRY_DISPUTED string = "XX"
const COUNTRY_UNKNOWN string = "XY"
const COUNTRY_COMPLICATED string = "XZ"

func Country(body []byte) string {

	rsp := gjson.GetBytes(body, "properties.wof:country")

	if !rsp.Exists() {
		return COUNTRY_UNKNOWN
	}

	return rsp.String()
}

func MergeCountries(features ...[]byte) string {

	tmp := make(map[string]bool)

	for _, body := range features {
		c := Country(body)
		tmp[c] = true
	}

	codes := make([]string, 0)

	for c, _ := range tmp {
		codes = append(codes, c)
	}

	switch len(codes) {
	case 0:
		return COUNTRY_UNKNOWN
	case 1:
		return codes[0]
	default:
		return COUNTRY_COMPLICATED
	}
}
