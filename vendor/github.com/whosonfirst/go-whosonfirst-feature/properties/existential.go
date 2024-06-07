package properties

import (
	"github.com/sfomuseum/go-edtf"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-flags/existential"
)

func IsCurrent(body []byte) (flags.ExistentialFlag, error) {

	rsp := gjson.GetBytes(body, "properties.mz:is_current")

	if rsp.Exists() {

		v := rsp.Int()

		if v == 1 || v == 0 {
			return existential.NewKnownUnknownFlag(v)
		}
	}

	d, err := IsDeprecated(body)

	if err != nil {
		return nil, err
	}

	if d.IsTrue() && d.IsKnown() {
		return existential.NewKnownUnknownFlag(0)
	}

	c, err := IsCeased(body)

	if err != nil {
		return nil, err
	}

	if c.IsTrue() && c.IsKnown() {
		return existential.NewKnownUnknownFlag(0)
	}

	s, err := IsSuperseded(body)

	if err != nil {
		return nil, err
	}

	if s.IsTrue() && s.IsKnown() {
		return existential.NewKnownUnknownFlag(0)
	}

	return existential.NewKnownUnknownFlag(-1)
}

func IsDeprecated(body []byte) (flags.ExistentialFlag, error) {

	rsp := gjson.GetBytes(body, "properties.edtf:deprecated")

	// "-" is not part of the EDTF spec it's just a default
	// string that we define for use in the switch statements
	// below (20210209/thisisaaronland)

	v := rsp.String()

	// 2019 EDTF spec (ISO-8601:1/2)

	switch v {
	case "-":
		return existential.NewKnownUnknownFlag(0)
	case edtf.UNKNOWN:
		return existential.NewKnownUnknownFlag(-1)
	default:
		// pass
	}

	// 2012 EDTF spec - annoyingly the semantics of ""
	// changed between the two (was meant to signal open
	// and now signals unknown)

	switch v {
	case "-":
		return existential.NewKnownUnknownFlag(0)
	case "u":
		return existential.NewKnownUnknownFlag(-1)
	case "uuuu":
		return existential.NewKnownUnknownFlag(-1)
	default:
		//
	}

	return existential.NewKnownUnknownFlag(1)
}

func IsCeased(body []byte) (flags.ExistentialFlag, error) {

	rsp := gjson.GetBytes(body, "properties.edtf:cessation")

	v := rsp.String()

	// 2019 EDTF spec (ISO-8601:1/2)

	switch v {
	case edtf.OPEN:
		return existential.NewKnownUnknownFlag(0)
	case edtf.UNKNOWN:
		return existential.NewKnownUnknownFlag(-1)
	default:
		// pass
	}

	// 2012 EDTF spec - annoyingly the semantics of ""
	// changed between the two (was meant to signal open
	// and now signals unknown)

	switch v {
	case "":
		return existential.NewKnownUnknownFlag(0)
	case "u":
		return existential.NewKnownUnknownFlag(-1)
	case "uuuu":
		return existential.NewKnownUnknownFlag(-1)
	default:
		// pass
	}

	return existential.NewKnownUnknownFlag(1)
}

func IsSuperseded(body []byte) (flags.ExistentialFlag, error) {

	rsp := gjson.GetBytes(body, "properties.wof:superseded_by")

	if rsp.Exists() && len(rsp.Array()) > 0 {
		return existential.NewKnownUnknownFlag(1)
	}

	return existential.NewKnownUnknownFlag(0)
}

func IsSuperseding(body []byte) (flags.ExistentialFlag, error) {

	rsp := gjson.GetBytes(body, "properties.wof:supersedes")

	if rsp.Exists() && len(rsp.Array()) > 0 {
		return existential.NewKnownUnknownFlag(1)
	}

	return existential.NewKnownUnknownFlag(0)
}
