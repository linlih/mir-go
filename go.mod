module mir-go

go 1.16

require (
	github.com/AlecAivazis/survey/v2 v2.2.12
	github.com/bluele/gcache v0.0.2 // indirect
	github.com/desertbit/grumble v1.1.1
	github.com/google/gopacket v1.1.19
	github.com/olekukonko/tablewriter v0.0.5
	github.com/panjf2000/ants v1.3.0
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/takama/daemon v1.0.0 // indirect
	github.com/urfave/cli/v2 v2.3.0
	gopkg.in/ini.v1 v1.62.0
	minlib v0.0.0
)

replace minlib v0.0.0 => ../minlib
