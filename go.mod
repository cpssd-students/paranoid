module paranoid

require (
	github.com/cheggaaa/pb v2.0.6+incompatible
	github.com/golang/protobuf v1.2.0
	github.com/hanwen/go-fuse v0.0.0-20181001113445-8ccb625b0adc
	github.com/huin/goupnp v0.0.0-20180415215157-1395d1447324
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1
	github.com/pp2p/paranoid v0.0.0-20170726085752-928192980d34
	github.com/urfave/cli v1.20.0
	golang.org/x/crypto v0.0.0-20180927165925-5295e8364332
	golang.org/x/net v0.0.0-20180926154720-4dfa2610cdf3
	google.golang.org/grpc v1.15.0
)

replace github.com/pp2p/paranoid => ./
