//&{Type:reaction_added User:U02E5H1L2 ItemUser:U02E5H1L2 Item:{Type:message Channel:C3TP862EP File: FileComment: Timestamp:1501303033.281065} Reaction:smile EventTimestamp:1501305117.417607}

Premise can only see slack user's token messages

First detect one kind of emoji was used in the message
then if an approved emoji get the callback information
in the callback info get message ID, if message ID exits in message list
we lookup app id and then we get the actual app object by id. We then 
pass the callback url and message token back to the application along with all the stuff

//hide the tokens in the get methods