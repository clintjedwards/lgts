# Snark

Snark is a service which monitors slack events for emoji reactions and then sends a request to the interested party.

## Summary

Traditionally, [interactive Slack messages](https://api.slack.com/docs/message-buttons) are supported via Slack's interactive button workflow. This requires the application that wants to process the action from the button click to publicly expose a callback URL that slack can `POST` back to when a button is clicked.

This is a problem for applications that want|need to stay completely internal, want to have interactivity, but still want to make use of Slack's awesome messaging platform.

In comes in LGTS. LGTS uses Slack's event stream API to monitor Slack channels that its a member of. When a message recieves an "emoji added" event it scans the messages for a preshared ID registered negotiated earlier with the relevant application. When the ID is found it checks its list of accepted emojis and messages in queue to make sure things adhere to the applications requirements for contact. Once it verifies this is a message that should be acted on, Snark sends a request back to the requesting application, enabling interactivity through emojis.

## Installation

```
go get -u github.com/clintjedwards/lgts
```

## Configuration

### Slack bot setup

You will need to setup a slack bot in order to get the API keys required to have Snark monitor messages

1.  Go to https://api.slack.com/apps
2.  Create a new app
3.  Create a new bot user
4.  Add the following permissions: ["channels.history", "reactions:read", "users:read.email"]
5.  Install app to workspace
6.  Save the two tokens somewhere safe, this is your app token and bot token respectively

### Environment Variables

Snark takes its configuration from environment variables. The following variables are required:

* `SERVER_URL`
* `SLACK_APP_TOKEN`
* `SLACK_BOT_TOKEN`

The `DEBUG` environment variable can be turned off and on to control the verbosity

## Usage

### **Step 1:** Register a message

The first step is to register a prospective message with snark first. You must provide the following things

* callback_url: This is a route to an app that snark will send the emoji events back to. Usually in the form
  `https://exampleapp.com/callback`.
* auth_token: This auth token will be passed back to the callback address as a way to verify that the message has
  come from snark.
* valid_emojis: A json list of valid emoji strings that snark will send an event for. All other emojis are ignored.

Example:

```json
#Request
$ http POST https://snark.server.example/track auth_token="somesecret" callback_url="myapp.example/callback" valid_emojis:='["lgts", "wut"]'

#Response
HTTP/1.1 201 Created
Access-Control-Allow-Origin: *
Content-Length: 43
Content-Type: application/json
Date: Sun, 13 May 2018 02:20:02 GMT

{
    "message_id": "d911cb2702245d470c4f"
}
```

The message ID should be stored by the calling application as a way to map events sent back to messages posted to slack.

### **Step 2:** Post your message

Next your app should post the slack message with an attachment. In the attachment field pass the "callback_id" with the value of the message ID return to you above.

Example:

```json
{
    "text": "Would you like to play a game?",
    "attachments": [
        {
            "text": "Choose a game to play",
            "fallback": "You are unable to choose a game",
            "callback_id": "d911cb2702245d470c4f",
            "color": "#3AA3E3",
            "attachment_type": "default",
            ...
```

### **Step 3:** Receive emoji events on the callback URL

Once you've posted the message snark will automatically alert you of any emojis that fit your criteria. You'll get the
following fields as part of a POST request sent to the callback URL you specified

* id: The message ID associated with the event.
* emoji_used: The text string of the emoji used.
* auth_token: The auth token associated with the message ID. You should check this to make sure the message came from snark
* slack_user_email: The slack user's email who sent the event.
* slack_user_name: The slack user's full name who sent the event.

### **Step 4:** Once you've recieved the event you're listening for unsubscribe from the message

You can do this by sending a delete request like so:

```json
#Request
$ http DELETE https://snark.server.example/track/d911cb2702245d470c4f auth_token="somesecret"

#Response
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: *
Content-Type: application/json
Date: Sun, 13 May 2018 03:30:35 GMT
```

## Authors

* **Clint Edwards** - [Github](https://github.com/clintjedwards)
