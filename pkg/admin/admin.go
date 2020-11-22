// Package admin provides an interface to connect to and manage a running
// OpenTTD dedicated server.
package admin

// references
// https://webcache.googleusercontent.com/search?q=cache:prtEeaJZJFQJ:https://wiki.openttd.org/Server_Admin_Port_Development+&cd=2&hl=en&ct=clnk&gl=au&client=safari
// https://github.com/OpenTTD/OpenTTD/blob/master/src/network/core/tcp_admin.h
// https://github.com/OpenTTD/OpenTTD/blob/master/docs/admin_network.md
// https://github.com/OpenTTD/OpenTTD/blob/master/src/network/network_admin.cpp

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// OpenTTDServer - an object representing the server connection
type OpenTTDServer struct {
	connection      net.Conn
	ServerName      string
	ServerVersion   string
	ServerDedicated bool // is this a dedicated server?
	MapName         string
	MapSeed         uint32
	MapLandscape    byte
	MapX            uint16
	MapY            uint16

	rconDaily    []string
	rconMonthly  []string
	rconYearly   []string
	connected    chan bool
	disconnected chan bool
}

const (
	adminPacketAdminJOIN             = 0 ///< The admin announces and authenticates itself to the server.
	adminPacketAdminQUIT             = 1 ///< The admin tells the server that it is quitting.
	adminPacketAdminUPDATE_FREQUENCY = 2 ///< The admin tells the server the update frequency of a particular piece of information.
	adminPacketAdminPOLL             = 3 ///< The admin explicitly polls for a piece of information.
	adminPacketAdminCHAT             = 4 ///< The admin sends a chat message to be distributed.
	adminPacketAdminRCON             = 5 ///< The admin sends a remote console command.
	adminPacketAdminGAMESCRIPT       = 6 ///< The admin sends a JSON string for the GameScript.
	adminPacketAdminPING             = 7 ///< The admin sends a ping to the server, expecting a ping-reply (PONG) packet.

	adminPacketServerFULL     = 100 ///< The server tells the admin it cannot accept the admin.
	adminPacketServerBANNED   = 101 ///< The server tells the admin it is banned.
	adminPacketServerERROR    = 102 ///< The server tells the admin an error has occurred.
	adminPacketServerPROTOCOL = 103 ///< The server tells the admin its protocol version.
	adminPacketServerWELCOME  = 104 ///< The server welcomes the admin to a game.
	adminPacketServerNEWGAME  = 105 ///< The server tells the admin its going to start a new game.
	adminPacketServerSHUTDOWN = 106 ///< The server tells the admin its shutting down.

	adminPacketServerDATE            = 107 ///< The server tells the admin what the current game date is.
	adminPacketServerCLIENT_JOIN     = 108 ///< The server tells the admin that a client has joined.
	adminPacketServerCLIENT_INFO     = 109 ///< The server gives the admin information about a client.
	adminPacketServerCLIENT_UPDATE   = 110 //< The server gives the admin an information update on a client.
	adminPacketServerCLIENT_QUIT     = 111 ///< The server tells the admin that a client quit.
	adminPacketServerCLIENT_ERROR    = 112 ///< The server tells the admin that a client caused an error.
	adminPacketServerCOMPANY_NEW     = 113 ///< The server tells the admin that a new company has started.
	adminPacketServerCOMPANY_INFO    = 114 ///< The server gives the admin information about a company.
	adminPacketServerCOMPANY_UPDATE  = 115 ///< The server gives the admin an information update on a company.
	adminPacketServerCOMPANY_REMOVE  = 116 ///< The server tells the admin that a company was removed.
	adminPacketServerCOMPANY_ECONOMY = 117 ///< The server gives the admin some economy related company information.
	adminPacketServerCOMPANY_STATS   = 118 ///< The server gives the admin some statistics about a company.
	adminPacketServerCHAT            = 119 ///< The server received a chat message and relays it.
	adminPacketServerRCON            = 120 ///< The server's reply to a remove console command.
	adminPacketServerCONSOLE         = 121 ///< The server gives the admin the data that got printed to its console.
	adminPacketServerCMD_NAMES       = 122 ///< The server sends out the names of the DoCommands to the admins.
	adminPacketServerCMD_LOGGING     = 123 ///< The server gives the admin copies of incoming command packets.
	adminPacketServerGAMESCRIPT      = 124 ///< The server gives the admin information from the GameScript in JSON.
	adminPacketServerRCON_END        = 125 ///< The server indicates that the remote console command has completed.
	adminPacketServerPONG            = 126 ///< The server replies to a ping request from the admin.

	invalidAdminPacket = 255 ///< An invalid marker for admin packets.

	adminUpdateDATE            = 0  ///< Updates about the date of the game.
	adminUpdateCLIENT_INFO     = 1  ///< Updates about the information of clients.
	adminUpdateCOMPANY_INFO    = 2  ///< Updates about the generic information of companies.
	adminUpdateCOMPANY_ECONOMY = 3  ///< Updates about the economy of companies.
	adminUpdateCOMPANY_STATS   = 4  ///< Updates about the statistics of companies.
	adminUpdateCHAT            = 5  ///< The admin would like to have chat messages.
	adminUpdateCONSOLE         = 6  ///< The admin would like to have console messages.
	adminUpdateCMD_NAMES       = 7  ///< The admin would like a list of all DoCommand names.
	adminUpdateCMD_LOGGING     = 8  ///< The admin would like to have DoCommand information.
	adminUpdateGAMESCRIPT      = 9  ///< The admin would like to have gamescript messages.
	adminUpdateEND             = 10 ///< Must ALWAYS be on the end of this list!! (period)

	adminFrequencyPOLL      = 0x01 ///< The admin can poll this.
	adminFrequencyDAILY     = 0x02 ///< The admin gets information about this on a daily basis.
	adminFrequencyWEEKLY    = 0x04 ///< The admin gets information about this on a weekly basis.
	adminFrequencyMONTHLY   = 0x08 ///< The admin gets information about this on a monthly basis.
	adminFrequencyQUARTERLY = 0x10 ///< The admin gets information about this on a quarterly basis.
	adminFrequencyANUALLY   = 0x20 ///< The admin gets information about this on a yearly basis.
	adminFrequencyAUTOMATIC = 0x40 ///< The admin gets information about this when it changes.
)

// Connect to the OpenTTD server on the admin port. Requires that the server
// is listening on the admin port (admin_password must be specified in the config).
// This method will block, and automatically attempt to reconnect if disconnected.
func (server *OpenTTDServer) Connect(host string, port int, password string, botName string, botVersion string) {

	for {

		// fmt.Printf("array: %v (%T) %d\n", toSend, toSend, size)
		log.Println("connecting...")
		connectString := fmt.Sprintf("%s:%d", host, port)
		conn, err := net.Dial("tcp", connectString)
		if err != nil {
			log.Printf("error connecting: %v\n", err)
			time.Sleep(time.Second * 2)
			continue
			//panic(err)
		}

		log.Println("connected")

		go server.listenSocket()

		server.connected = make(chan bool)
		server.disconnected = make(chan bool)

		server.connection = conn
		server.connected <- true

		// start listening

		// login
		var toSend []byte
		toSend = append(toSend[:], adminPacketAdminJOIN) // type
		toSend = append(toSend[:], []byte(password)...)  // password
		toSend = append(toSend[:], 0x0)
		toSend = append(toSend[:], []byte(botName)...) // client name
		toSend = append(toSend[:], 0x0)
		toSend = append(toSend[:], []byte(botVersion)...) // version
		toSend = append(toSend[:], 0x0)
		size := len(toSend) + 2

		toSend = append([]byte{byte(size), 0x0}, toSend[:]...)
		server.connection.Write(toSend)

		// register for daily updates
		updateDateCmd := make([]byte, 2)
		binary.LittleEndian.PutUint16(updateDateCmd, adminUpdateDATE)
		updateDateDaily := make([]byte, 2)
		binary.LittleEndian.PutUint16(updateDateDaily, adminFrequencyDAILY)

		toSend = []byte{}
		toSend = append(toSend, updateDateCmd...)
		toSend = append(toSend, updateDateDaily...)
		server.sendSocket(adminPacketAdminUPDATE_FREQUENCY, toSend)

		// toSend = []byte{}
		// toSend = append(toSend[:], adminPacketAdminUPDATE_FREQUENCY)
		// toSend = append(toSend[:], adminUpdateCHAT, 0x0)
		// toSend = append(toSend[:], adminFrequencyAUTOMATIC, 0x0)

		// size = len(toSend) + 2
		//
		// toSend = append([]byte{byte(size), 0x0}, toSend[:]...)
		// fmt.Printf("array: %v (%T) %d\n", toSend, toSend, size)
		// conn.Write(toSend)

		// wait until we are told we disconnected
		<-server.disconnected
		fmt.Printf("Reconnecting....")
		time.Sleep(2 * time.Second)
	}

}

// RegisterDateChange sends an arbitrary number of rcon commands when certain
// (game time) date changes occur. The possible periods are 'daily', 'monthly'
// and 'yearly'.
//
// Note that this must be called before Connect.
func (server *OpenTTDServer) RegisterDateChange(period string, command string) {
	if period == "daily" {
		server.rconDaily = append(server.rconDaily, command)
	} else if period == "monthly" {
		server.rconMonthly = append(server.rconMonthly, command)
	} else if period == "yearly" {
		server.rconYearly = append(server.rconYearly, command)
	} else {
		panic("bad period " + period)
	}
	return
}

func (server *OpenTTDServer) dateChanged(dt time.Time) {
	// do every daily one
	for _, rconCommand := range server.rconDaily {
		server.rconCommand(processCommand(rconCommand, dt))
	}

	// monthly ones on the 1st
	if dt.Day() == 1 {
		for _, rconCommand := range server.rconMonthly {
			server.rconCommand(processCommand(rconCommand, dt))
		}
	}

	// yearly on the 1st of jan
	if dt.Day() == 1 && dt.Month() == 1 {
		for _, rconCommand := range server.rconYearly {
			server.rconCommand(processCommand(rconCommand, dt))
		}
	}

}

func (server OpenTTDServer) rconCommand(command string) {

	var rconCommand []byte
	rconCommand = append(rconCommand, command...)
	rconCommand = append(rconCommand, 0000)

	server.sendSocket(adminPacketAdminRCON, rconCommand)
}

func processCommand(command string, dt time.Time) string {
	command = strings.Replace(command, "%Y", fmt.Sprintf("%04d", dt.Year()), -1)
	command = strings.Replace(command, "%M", fmt.Sprintf("%02d", dt.Month()), -1)
	command = strings.Replace(command, "%D", fmt.Sprintf("%02d", dt.Day()), -1)
	return command
}

func (server *OpenTTDServer) sendSocket(protocol int, data []byte) {
	// fmt.Printf("Going to send using protocol %v this data: %v\n", protocol, data)
	toSend := make([]byte, 3)     // start with 3 bytes for the length and protocol
	size := uint16(len(data) + 3) // size 2 bytes, plus protocol
	binary.LittleEndian.PutUint16(toSend, size)
	// toSend = append(toSend[:],
	toSend[2] = byte(protocol)
	toSend = append(toSend, data...)
	// fmt.Printf("Going to send this: %v\n", toSend)
	server.connection.Write(toSend)
}

func (server *OpenTTDServer) listenSocket() {

	// fmt.Println("waiting for connection...")
	// fmt.Printf("Listening to socket...\n")
	<-server.connected

	var chunk []byte

SocketLoop:
	for {

		// fmt.Printf("Waiting for socket data\n")
		socketData := make([]byte, 1024)
		s, err := server.connection.Read(socketData)
		if err != nil {
			if cErr, ok := err.(*net.OpError); ok {
				if cErr.Err.Error() == "read: connection reset by peer" {
					log.Println("Connection reset by peer - check the openttd log for details")
					server.connection = nil
					server.disconnected <- true
					return

				}
			} else {
				log.Println("Error occurred on socket: ", err)
				server.connection = nil
				server.disconnected <- true
				return
			}
			return
		}

		// fmt.Printf("Read %d bytes from socket\n", s)

		// append this data to the chunk that's in progress
		chunk = append(chunk, socketData[0:s]...)

		for {

			// do we have enough data to process?
			// which means at least 2 bytes for the length header
			if len(chunk) < 2 {
				continue SocketLoop
			}

			// read the packet size at the front
			packetSize := binary.LittleEndian.Uint16(chunk[0:2])

			// if we don't have enough bytes yet, just loop around
			// fmt.Printf("current chunk %d bytes, indicated packet size %d\n", len(chunk), packetSize)
			if packetSize > uint16(len(chunk)) {
				// fmt.Printf("incomplete data, waiting for more from socket\n")
				continue SocketLoop
			}

			// so we are good to continue processing
			packetType := int(chunk[2])
			packetData := chunk[3:packetSize]
			// fmt.Printf("packet type %d and size is %v bytes, I read %d from socket\n", packetType, len(packetData), s)

			if packetType == adminPacketServerPROTOCOL {
				// fmt.Print(" - Got a adminPacketServerPROTOCOL packet\n")
			} else if packetType == adminPacketServerWELCOME {
				log.Println("received welcome packet")
				fmt.Printf("packet: %+v?\n", packetData[:])
				var next int
				server.ServerName, next = extractString(packetData[:], 0)
				server.ServerVersion, next = extractString(packetData[:], next)
				if packetData[next] == 0000 {
					server.ServerDedicated = false
				} else if packetData[next] == 0001 {
					server.ServerDedicated = true
				} else {
					fmt.Printf("not bool %v?\n", packetData[next])
				}
				server.MapName, next = extractString(packetData[:], next+1)
				server.MapSeed = binary.LittleEndian.Uint32(packetData[next : next+4])
				server.MapLandscape = packetData[next+4]
				// todo
				// p->Send_uint32(ConvertYMDToDate(_settings_game.game_creation.starting_year, 0, 1));
				server.MapX = binary.LittleEndian.Uint16(packetData[next+9 : next+11])
				server.MapY = binary.LittleEndian.Uint16(packetData[next+11 : next+13])

				log.Printf("server: %s version: %s dedicated: %v map: %s %d/%d size\n", server.ServerName, server.ServerVersion, server.ServerDedicated, server.MapName, server.MapX, server.MapY)

				// fmt.Printf("   * server name: %s\n", ServerName)
			} else if packetType == adminPacketServerSHUTDOWN {
				log.Println("server shutting down - will try to reconnect")
				server.connection = nil
				server.disconnected <- true
				return

			} else if packetType == adminPacketServerDATE {
				// [[7 0 107 84 252 10 0 0 0
				date := binary.LittleEndian.Uint32(packetData[0:4])
				epochDate := time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC)
				dt := epochDate.AddDate(0, 0, int(date))
				// fmt.Printf("   * Date is %v\n", dt)
				server.dateChanged(dt)
				// uint32
			} else if packetType == adminPacketServerCHAT {
				// fmt.Printf(" - Got a chat packet:\n%v", packetData)
				// [3 0 3 0 0 0 98 105 116 104 99 105 110 103 0 0 0 0 0 0 0 0 0]
				chatAction := int8(packetData[0])
				chatDestType := int8(packetData[1])
				chatClientID := binary.LittleEndian.Uint32(packetData[2:6])
				// var chatMsg string
				chatMsg, _ := extractString(packetData[:], 6)
				chatData := binary.LittleEndian.Uint64(packetData[len(packetData)-8:])
				log.Printf("chat message: action %v desttype %v, client id %v msg %v data %v\n", chatAction, chatDestType, chatClientID, string(chatMsg), chatData)
			} else if packetType == adminPacketServerRCON {
				colour := binary.LittleEndian.Uint16(packetData[0:2])
				rconRecvString, _ := extractString(packetData[:], 2)
				log.Printf("rcon: colour %v : %s\n", colour, rconRecvString)
			} else if packetType == adminPacketServerRCON_END {
				rconEndRecvString, _ := extractString(packetData[:], 0)
				log.Printf("rcon end : %s\n", rconEndRecvString)
			} else {
				log.Printf("unknown packet received from server: %v [%v]\n", string(packetData), packetData)
			}

			// fmt.Printf("removing the chunk we have processed\n")
			chunk = chunk[packetSize:]
		}

		// check if there is data left to process in the current data
		// fmt.Printf("remaining in chunk %d bytes", len(chunk))
		if len(chunk) < 3 {
			// we don't even have enough for a length and protocol type, so may
			// as well go sit on the socket
			continue SocketLoop
		}

	}

}

func extractString(bytes []byte, start int) (string, int) {
	var buildString []byte
	for i := start; i <= len(bytes); i++ {
		if bytes[i] == 0 {
			return string(buildString), i + 1
		}
		buildString = append(buildString, bytes[i])
	}
	return "", -1
}
