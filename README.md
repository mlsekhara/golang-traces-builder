# golang-traces-builder

Sample golang script to build a nice json output tree that contains the trace of communication between different services inside a microservices ecosystem.


# Input

2016-10-20T12:43:34.000Z 2016-10-20T12:43:35.000Z t1 s-3 ab->ad
2016-10-20T12:43:31.000Z 2016-10-20T12:43:40.000Z t1 s-1 ad->ac
2016-10-20T12:43:39.000Z 2016-10-20T12:43:40.000Z t1 s-2 aa->ab
2016-10-20T12:43:33.000Z 2016-10-20T12:43:42.000Z t1 front-end null->aa

# Ouput

```json
{
  "id": "t1",
  "root": {
    "service": "front-end",
    "start": "2016-10-20T12:43:33.000Z",
    "end": "2016-10-20T12:43:42.000Z",
    "calls": [
      {
        "service": "s-2",
        "start": "2016-10-20T12:43:39.000Z",
        "end": "2016-10-20T12:43:40.000Z",
        "calls": [
          {
            "service": "s-3",
            "start": "2016-10-20T12:43:34.000Z",
            "end": "2016-10-20T12:43:35.000Z",
            "calls": [
              {
                "service": "s-1",
                "start": "2016-10-20T12:43:31.000Z",
                "end": "2016-10-20T12:43:40.000Z",
                "calls": [],
                "span": "ac"
              }
            ],
            "span": "ad"
          }
        ],
        "span": "ab"
      }
    ],
    "span": "aa"
  }
}
```
