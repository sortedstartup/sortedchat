Options: viper vs koanf

koanf is maintained by knadh (zerodha)
I have been using viper earlier but it has issues
https://github.com/knadh/koanf/wiki/Comparison-with-spf13-viper

1. forces lower case keys ( can be set in configured to become case sensitive)
2. increases binary size by 10 MB at least compare to koanf and plain json
3. not designed to be extended easily
4. large 3rd part dependencies
5. Get() returns refs which can be modified which modified underlying configuration

