# run with python3 ./twitter.py 
# This is callable from Go as follows:
#   python3 ./twitter-post-message.py <TweetID> <message>
# eg: 
#   python3 ./twitter-post-message.py 1795574545480286466 "ready for this?"
# prints out a json body on execution:
# EG:
# {
#     "data": {
#         "edit_history_tweet_ids": [
#             "1795574718277279780"
#         ],
#         "id": "1795574718277279780",
#         "text": "ready for this?"
#     }
# }
from requests_oauthlib import OAuth1Session
import os
import json
import argparse



parser = argparse.ArgumentParser(description='post a tweet reply')
parser.add_argument('tweetid', type=str, nargs='*',
                    help='ID of the tweet to reply to')
parser.add_argument('message', type=str, nargs='+',
                    help='message to reply to that tweet')
args = parser.parse_args()


tweetID = args.tweetid #"1795570508680823239"
message = args.message[0]

# this is our API key and secret - not our oauth one!
consumer_key = os.environ.get("TWITTER_API_KEY")
consumer_secret = os.environ.get("TWITTER_API_KEY_SECRET")

# Be sure to add replace the text of the with the text you wish to Tweet. You can also add parameters to post polls, quote Tweets, Tweet with reply settings, and Tweet to Super Followers in addition to other features.
payload = {"text": message}
if len(tweetID) >0:
    payload = {"text": message , "reply": { "in_reply_to_tweet_id": tweetID[0] }}

access_token = os.environ.get("TWITTER_USER_TOKEN")
access_token_secret = os.environ.get("TWITTER_USER_SECRET")

# Make the request
oauth = OAuth1Session(
    consumer_key,
    client_secret=consumer_secret,
    resource_owner_key=access_token,
    resource_owner_secret=access_token_secret,
)

# Making the request
response = oauth.post(
    "https://api.twitter.com/2/tweets",
    json=payload,
)

if response.status_code != 201:
    raise Exception(
        "Request returned an error: {} {}".format(response.status_code, response.text)
    )


# Saving the response as JSON
json_response = response.json()
print(json.dumps(json_response, indent=4, sort_keys=True))

# {
#     "data": {
#         "edit_history_tweet_ids": [
#             "1795570508680823239"
#         ],
#         "id": "1795570508680823239",
#         "text": "another test or 4"
#     }
# }