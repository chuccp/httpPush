package core

import "errors"

var NoFoundUser = errors.New("NOT FOUND USER")

var UnKnownMessageType = errors.New("UNKNOWN MESSAGE TYPE")

var UnKnownClassId = errors.New("UNKNOWN CLASS ID")

var MessageIdIsBlank = errors.New("message Id is blank")

var ProtocolError = errors.New("km protocolError")

var ReadTimeout = errors.New("read timeout")

var ConnOnCreating = errors.New("conn  on creating ")

var UnKnownConn = errors.New("UNKNOWN Conn")

var UnKnownVersion = errors.New("UNKNOWN Version")

var NetError = errors.New("net ERROR")
