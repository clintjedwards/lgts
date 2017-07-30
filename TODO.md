1) Application registers with lgts
2) Application posts to lgts message route with a pre-generated random key before every message. Each key sent per message is unique
3) Application then posts message to slack with any extra details of the message's content it wants back in the callback_id of the message(json format)
4) lgts parses all messages in slack and in certain channels and when a reaction type is read compares it against its database of registered messages
5) If it finds a message that matches the message_id or slack user it sends it back to the callback id for the messsage_id or slack user
6) The application handles this how ever it wants


* Change all return message to something consistent
* Finish implementing proper slack workflow and calls
* 

//&{Type:reaction_added User:U02E5H1L2 ItemUser:U02E5H1L2 Item:{Type:message Channel:C3TP862EP File: FileComment: Timestamp:1501303033.281065} Reaction:smile EventTimestamp:1501305117.417607}
