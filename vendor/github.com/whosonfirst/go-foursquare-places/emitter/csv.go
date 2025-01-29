package emitter

import (
	"compress/bzip2"
	"context"
	"io"
	"iter"
	_ "log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/sfomuseum/go-csvdict"
	"github.com/whosonfirst/go-foursquare-places"
)

type CSVEmitter struct {
	Emitter
	reader       io.ReadCloser
	bzip2_reader io.Reader
}

func init() {

	ctx := context.Background()
	err := RegisterEmitter(ctx, "csv", NewCSVEmitter)

	if err != nil {
		panic(err)
	}
}

func NewCSVEmitter(ctx context.Context, uri string) (Emitter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	r, err := os.Open(u.Path)

	if err != nil {
		return nil, err
	}

	br := bzip2.NewReader(r)

	e := &CSVEmitter{
		reader:       r,
		bzip2_reader: br,
	}

	return e, nil
}

func (e *CSVEmitter) Emit(ctx context.Context) iter.Seq2[*places.Place, error] {

	return func(yield func(*places.Place, error) bool) {

		csv_r, err := csvdict.NewReader(e.bzip2_reader)

		if err != nil {
			yield(nil, err)
			return
		}

		for {

			row, err := csv_r.Read()

			if err == io.EOF {
				break
			}

			if err != nil {
				yield(nil, err)
				continue
			}

			lat, err := strconv.ParseFloat(row["latitude"], 64)

			if err != nil {
				// slog.Warn("Failed to parse latitude", "id", row["fsq_place_id"], "name", row["name"], "latitude", row["latitude"], "error", err)
				lat = 0.0
			}

			lon, err := strconv.ParseFloat(row["longitude"], 64)

			if err != nil {
				// slog.Warn("Failed to parse longitude", "id", row["fsq_place_id"], "name", row["name"], "longitude", row["longitude"], "error", err)
				lon = 0.0
			}

			pl := &places.Place{
				Id:            row["fsq_place_id"],
				Name:          row["name"],
				Address:       row["address"],
				DateClosed:    row["date_closed"],
				DateCreated:   row["date_created"],
				DateRefreshed: row["date_refreshed"],
				Email:         row["email"],
				FacebookId:    row["facebook_id"],
				Instagram:     row["instagram"],
				PostBox:       row["po_box"],
				PostTown:      row["post_town"],
				PostCode:      row["post_code"],
				Region:        row["region"],
				Telephone:     row["telephone"],
				Twitter:       row["twitter"],
				Website:       row["website"],
				Country:       row["country"],
				Latitude:      lat,
				Longitude:     lon,
			}

			categories := make([]places.Category, 0)

			str_category_ids := row["fsq_category_ids"]
			str_category_ids = strings.TrimLeft(str_category_ids, "[")
			str_category_ids = strings.TrimLeft(str_category_ids, "]")

			str_category_labels := row["fsq_category_labels"]
			str_category_labels = strings.TrimLeft(str_category_labels, "[")
			str_category_labels = strings.TrimLeft(str_category_labels, "]")

			category_ids := strings.Split(str_category_ids, ", ")
			category_labels := strings.Split(str_category_labels, ", ")

			if len(category_ids) == len(category_labels) {

				for i, id := range category_ids {

					c := places.Category{
						Id:     id,
						Labels: strings.Split(category_labels[i], " > "),
					}

					categories = append(categories, c)
				}

			} else {

				// slog.Info("C", "c", category_ids)
				// slog.Info("C", "l", row["fsq_category_labels"])

				for _, id := range category_ids {

					c := places.Category{
						Id: id,
						// Labels: strings.Split(category_labels[i], " > "),
					}

					categories = append(categories, c)
				}

			}

			pl.Categories = categories

			yield(pl, nil)
		}
	}
}

func (e *CSVEmitter) Close() error {
	return e.reader.Close()
}
