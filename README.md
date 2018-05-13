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

## Usage

## Example

## Troubleshooting

## Authors

* **Clint Edwards** - [Github](https://github.com/clintjedwards)
