package places

import (
	"fmt"
)

// 2024/11/21 19:33:43 INFO ROW row="map[address:Mickiewicza 8 admin_region: country:PL date_closed: date_created:2015-05-06 date_refreshed:2024-06-27 dt:2024-11-19 email:teresa.glowacka.fotos.kolo@neostrada.pl facebook_id: fsq_category_ids:[4d4b7105d754a06378d81259] fsq_category_labels:[Retail] fsq_place_id:cb57d89eed29405b908b0b6e instagram: latitude:52.19266718928583 locality:Koło longitude:18.63343577621856 name:Fotos. Zakład fotograficzny. Głowacka T. po_box: post_town: postcode:62-600 region:Wielkopolskie tel:63 272 08 68 twitter: website:]

type Place struct {
	Id            string     `json:"fsq_place_id"`
	Country       string     `json:"country"`
	Address       string     `json:"address"`
	AdminRegion   string     `json:"admin_region"`
	DateClosed    string     `json:"date_closed"`
	DateCreated   string     `json:"date_created"`
	DateRefreshed string     `json:"date_refreshed"`
	Email         string     `json:"email"`
	FacebookId    string     `json:"facebook_id"`
	Instagram     string     `json:"instagram"`
	Latitude      float64    `json:"latitude"`
	Longitude     float64    `json:"longitude"`
	Name          string     `json:"name"`
	PostBox       string     `json:"po_box"`
	PostTown      string     `json:"post_town"`
	PostCode      string     `json:"post_code"`
	Region        string     `json:"region"`
	Telephone     string     `json:"tel"`
	Twitter       string     `json:"twitter"`
	Website       string     `json:"website"`
	Categories    []Category `json:"categories"`
}

func (pl *Place) String() string {
	return fmt.Sprintf("%s %s", pl.Name, pl.Id)
}

type Category struct {
	Id     string   `json:"id"`
	Labels []string `json:"labels"`
}
