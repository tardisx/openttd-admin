# openttd-admin

This is a Golang interface to an OpenTTD server, and a standalone 'openttd_multitool' binary
for simple, periodic server operations.

The latter might include periodically:

* sending messages to connected players
* saving the game to a custom, datestamped save
* generating datestamped screenshots

## admin.go library

At the moment the library has limited support for anything except managing
responses to date changes. Please tell me your use case and I will be happy
to look at extending it further. Or, pull requests accepted :-)

## openttd_multitool

The openttd_multitool connects to the OpenTTD Admin port (default 3977)
and stays connected. It monitors the game date, and performs your custom commands
at periodic intervals.

Possible intervals are:

* daily
* monthly
* yearly

These intervals are obviously in game time!

## running

The tool is a command line driven executable. It does not require installation.
Just copy it somewhere and run it.

You can configure the tool to send any command that you would type at the OpenTTD
console. Here are a few examples:

    openttd_multitool -hostname localhost -password abc -daily "say \"Hello to all players\""

This sends an annoying message to all players, every day. Note the quoting requirements to ensure
that the argument to the "say" command in OpenTTD comes through as a single argument.

    openttd_multitool -hostname localhost -password abc -monthly "save mygame-%Y-%M"

This saves the game once per month, with a filename like "mygame-2020-11.sav".

    openttd_multitool -hostname localhost -password abc -yearly "screenshot giant screenshot-%Y%M%D"

This generates a screenshot once per year of the entire map, with a name like "screenshot-20201121.png".

NOTE that your OpenTTD server needs to support generating screenshots (some dedicated servers compiled without
graphics will not work) and the appropriate graphics packs need to also be installed.

Additionally, when using the "screenshot giant" command, the entire server will freeze for that time, almost
certainly kicking off all clients, unless your map is very small or your server is very very fast!

No harm is done in that case, they can just reconnect after the screenshot is finished.
