
# overview - initial idea

1) polling lamda function looking for tweets
    1a) push possible tweets to sns

2) triggered lamda function from incoming sns message
    2a) review
    2b) if good push reply to sns queue for the right responder

3) triggered lamda function from outgoing sns message
    2a) post reply based on message contents

## questions

- how do we manage creds in lamda?
  - KMS, then decrypt in code: https://openupthecloud.com/kms-aws-lambda/

- can we push to sns from lamda?  how? what permissions are needed?

## TODO

Cam to run through this tutorial & fully deploy it as-is to understand lamda & permissions:
https://www.alexedwards.net/blog/serverless-api-with-go-and-aws-lambda


