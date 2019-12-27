# StreamQL

## NOT READY FOR PRODUCTION

StreamQL is a proof of concept approach for directly streaming data from sql to an io.Writer so that you never need to materialize the results fully into memory.

It allows you to also bind each row to a struct, so that you can then more easily modify and enrich the data via an Encoder function, before finally being
written to the writer.

## Tips

You likely want your Encoder function to be some kind of closure over other services or objects to assist in enriching the struct before you ultimately marshal it into raw bytes for the writer, but also you may not need to do that, and any function with a signature of (interface{}) ([]byte, error) will work just fine.

## Note

This is so untested, I literally have not even tried it against a real db driver, just some mocks in tests to try and figure out some parts of the reflect package. But if the real world is anything like my tests (highly unlikely), it should be somewhat functional.

It's also designed to try and make your life a little easier w/r/t nils, nulls, and struct fields that may be a value, or may be a pointer. That part is probably where I'm the riskiest w/r/t the reflect behavior I'm using but hey, that's what learning experiences are all about.

## TODO
* Actually test it
* Better docs