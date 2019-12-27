package streamql

import (
	"errors"
	"io"
	"reflect"

	"github.com/fatih/structtag"
)

// Encoder is a function that turns an object into bytes, optionally
// returning an error if needed
type Encoder func(i interface{}) ([]byte, error)

// NextColsScanner is an interface that only deals with the few functions from sql.Rows
// that it needs, allowing for easier mocking/testing of Stream
type NextColsScanner interface {
	Next() bool
	Scan(...interface{}) error
	Columns() ([]string, error)
}

// Stream is a function which takes in sql.Rows, a destination-example pointer, an encoder to turn
// the sql results into bytes with, and then finally a writer to write to. The purpose is to
// streamline processing large datasets into a struct, then further into a format for another
// process (ie: json, yaml, etc) such that you could use a closure to make encoder more of a
// "transform+encode", which would allow you to insert other business logic to hydrate the struct
// before turning it into bytes. Long winded way of saying: Stream hydrated objects off of large
// sql results to a writer, likely a response body without having to hold the whole result set
// in memory
func Stream(rows NextColsScanner, dst interface{}, enc Encoder, wc io.Writer) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer to a struct, not a struct")
	}
	// Now that we know it's a ptr, get the element to actually reflect
	rv = rv.Elem()
	// Get the type need it for tag parsing
	t := reflect.TypeOf(dst).Elem()

	fc := len(cols)
	tagToField := make(map[string]string, fc)

	for i := 0; i < fc; i++ {
		tf := t.Field(i)
		tags, err := structtag.Parse(string(tf.Tag))
		if err != nil {
			return err
		}
		tv, err := tags.Get("db")
		if err != nil {
			return err
		}
		tagToField[tv.Name] = tf.Name
	}
	for rows.Next() {
		// Make a new slice of interfaces to scan into
		// TODO tried to make this directly the list of fields in-order of cols
		// on `dst` but it wouldn't populate the fields. Likely due to a complete
		// proper lack of understanding of how reflect works in go
		fields := make([]interface{}, fc)
		if err := rows.Scan(fields...); err != nil {
			return err
		}
		for i := 0; i < fc; i++ {
			// Get the field with the tag name that matches the column
			// and then set the value to the reflect value of the field
			// at the same ordinal as the column
			// TODO will obviously crash for a bunch of reasons such as
			// not dealing with ptrs, nils, etcs
			fv := reflect.ValueOf(fields[i])
			trv := rv.FieldByName(tagToField[cols[i]])
			// In the event that the struct wants a ptr but the row-cols have a non-ptr...
			if trv.Kind() == reflect.Ptr && fv.Kind() != reflect.Ptr {
				// first, set it to an intiialized zero value for it's underlying type
				// This allows us to set it in the ptr on the struct later
				trv.Set(reflect.New(trv.Type().Elem()))
				trv = trv.Elem()
			}
			trv.Set(fv)
		}
		// Pass it through the provided encoder so it can be turned to bytes
		b, err := enc(rv.Interface())
		if err != nil {
			return err
		}
		// Now, begin the write loop
		w := 0 // track total bytes written
		// keep going until we've written every byte
		for w < len(b) {
			n, err := wc.Write(b[w:]) // track total bytes written this attempt
			if err != nil {           // bail if err
				return err
			}
			w += n // add total bytes written this time to overall total
		}
	}
	return nil
}
