import base64
import json
import os
import random
import urllib.request
import uuid
from enum import Enum
from time import time
from typing import Union

import aiohttp

_PROXY = urllib.request.getproxies().get("https")

_BASE_OPTION_SETS = [
    "nlu_direct_response_filter",
    "deepleo",
    "disable_emoji_spoken_text",
    "responsible_ai_policy_235",
    "enablemm",
    "iycapbing",
    "iyxapbing",
    "dv3sugg",
    "iyoloxap",
    "iyoloneutral",
    "gencontentv3",
    "nojbfedge",
]


class _OptionSets(Enum):
    CREATIVE = _BASE_OPTION_SETS + ["h3imaginative"]
    BALANCED = _BASE_OPTION_SETS + ["galileo"]
    PRECISE = _BASE_OPTION_SETS + ["h3precise"]


_SLICE_IDS = [
    "winmuid1tf",
    "newmma-prod",
    "imgchatgptv2",
    "tts2",
    "voicelang2",
    "anssupfotest",
    "emptyoson",
    "tempcacheread",
    "temptacache",
    "ctrlworkpay",
    "winlongmsg2tf",
    "628fabocs0",
    "531rai268s0",
    "602refusal",
    "621alllocs0",
    "621docxfmtho",
    "621preclsvn",
    "330uaug",
    "529rweas0",
    "0626snptrcs0",
    "619dagslnv1nr"
]


class _LocationHint(Enum):
    USA = {
        "locale": "en-US",
        "LocationHint": [
            {
                "country": "United States",
                "state": "California",
                "city": "Los Angeles",
                "timezoneoffset": 8,
                "countryConfidence": 8,
                "Center": {
                    "Latitude": 34.0536909,
                    "Longitude": -118.242766,
                },
                "RegionType": 2,
                "SourceType": 1,
            },
        ],
    }
    CHINA = {
        "locale": "zh-CN",
        "LocationHint": [
            {
                "country": "China",
                "state": "",
                "city": "Beijing",
                "timezoneoffset": 8,
                "countryConfidence": 8,
                "Center": {
                    "Latitude": 39.9042,
                    "Longitude": 116.4074,
                },
                "RegionType": 2,
                "SourceType": 1,
            },
        ],
    }
    EU = {
        "locale": "en-IE",
        "LocationHint": [
            {
                "country": "Norway",
                "state": "",
                "city": "Oslo",
                "timezoneoffset": 1,
                "countryConfidence": 8,
                "Center": {
                    "Latitude": 59.9139,
                    "Longitude": 10.7522,
                },
                "RegionType": 2,
                "SourceType": 1,
            },
        ],
    }
    UK = {
        "locale": "en-GB",
        "LocationHint": [
            {
                "country": "United Kingdom",
                "state": "",
                "city": "London",
                "timezoneoffset": 0,
                "countryConfidence": 8,
                "Center": {
                    "Latitude": 51.5074,
                    "Longitude": -0.1278,
                },
                "RegionType": 2,
                "SourceType": 1,
            },
        ],
    }


_DELIMITER = '\x1e'
_FORWARDED_IP = f"1.0.0.{random.randint(0, 255)}"

_ALLOWED_MESSAGE_TYPES = [
    "ActionRequest",
    "Chat",
    "Context",
    "InternalSearchQuery",
    "InternalSearchResult",
    "Disengaged",
    "InternalLoaderMessage",
    "Progress",
    "RenderCardRequest",
    "AdsQuery",
    "SemanticSerp",
    "GenerateContentQuery",
    "SearchQuery"
]

_HEADERS = {
    "accept": "application/json",
    "accept-language": "en-US,en;q=0.9",
    "content-type": "application/json",
    "sec-ch-ua": '"Not_A Brand";v="99", Microsoft Edge";v="110", "Chromium";v="110"',
    "sec-ch-ua-arch": '"x86"',
    "sec-ch-ua-bitness": '"64"',
    "sec-ch-ua-full-version": '"109.0.1518.78"',
    "sec-ch-ua-full-version-list": '"Chromium";v="110.0.5481.192", "Not A(Brand";v="24.0.0.0", "Microsoft Edge";v="110.0.1587.69"',
    "sec-ch-ua-mobile": "?0",
    "sec-ch-ua-model": "",
    "sec-ch-ua-platform": '"Windows"',
    "sec-ch-ua-platform-version": '"15.0.0"',
    "sec-fetch-dest": "empty",
    "sec-fetch-mode": "cors",
    "sec-fetch-site": "same-origin",
    "x-ms-client-request-id": str(uuid.uuid4()),
    "x-ms-useragent": "azsdk-js-api-client-factory/1.0.0-beta.1 core-rest-pipeline/1.10.0 OS/Win32",
    "Referer": "https://www.bing.com/search?q=Bing+AI&showconv=1&FORM=hpcodx",
    "Referrer-Policy": "origin-when-cross-origin",
    "x-forwarded-for": _FORWARDED_IP,
}

_HEADERS_INIT_CONVER = {
    "authority": "www.bing.com",
    "accept": "application/json",
    "accept-language": "en-US,en;q=0.9",
    "cache-control": "max-age=0",
    "sec-ch-ua": '"Chromium";v="110", "Not A(Brand";v="24", "Microsoft Edge";v="110"',
    "sec-ch-ua-arch": '"x86"',
    "sec-ch-ua-bitness": '"64"',
    "sec-ch-ua-full-version": '"110.0.1587.69"',
    "sec-ch-ua-full-version-list": '"Chromium";v="110.0.5481.192", "Not A(Brand";v="24.0.0.0", "Microsoft Edge";v="110.0.1587.69"',
    "sec-ch-ua-mobile": "?0",
    "sec-ch-ua-model": '""',
    "sec-ch-ua-platform": '"Windows"',
    "sec-ch-ua-platform-version": '"15.0.0"',
    "upgrade-insecure-requests": "1",
    "user-agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36 Edg/112.0.1722.46",
    "x-edge-shopping-flag": "1",
    "x-forwarded-for": _FORWARDED_IP,
}


def _format(msg: dict) -> str:
    return json.dumps(msg, ensure_ascii=False) + _DELIMITER


async def create_conversation(
        proxy: str | None = _PROXY,
        cookies: list[dict] | None = None,
):
    formatted_cookies = {}
    if cookies:
        for cookie in cookies:
            formatted_cookies[cookie["name"]] = cookie["value"]
    async with aiohttp.ClientSession(
            cookies=formatted_cookies,
            headers=_HEADERS_INIT_CONVER,
    ) as session:
        response = await session.get(
            url="https://edgeservices.bing.com/edgesvc/turing/conversation/create",
            proxy=proxy,
        )
    if response.status != 200:
        text = await response.text()
        raise Exception(f"Authentication failed {text}")
    try:
        conversation = await response.json()
    except:
        text = await response.text()
        raise Exception(text)
    if conversation["result"]["value"] == "UnauthorizedRequest":
        raise Exception(conversation["result"]["message"])
    return conversation


def _get_location_hint_from_locale(locale: str) -> Union[dict, None]:
    locale = locale.lower()
    if locale == "en-gb":
        hint = _LocationHint.UK.value
    elif locale == "en-ie":
        hint = _LocationHint.EU.value
    elif locale == "zh-cn":
        hint = _LocationHint.CHINA.value
    else:
        hint = _LocationHint.USA.value
    return hint.get("LocationHint")


async def ask_stream(
        conversation: dict,
        prompt: str,
        context: str,
        conversation_style: str = "creative",
        locale: str = "zh-CN",
        proxy=_PROXY,
        image_url=None,
):
    timeout = aiohttp.ClientTimeout(total=900)
    async with aiohttp.ClientSession(timeout=timeout) as session:
        conversation_id = conversation["conversationId"]
        client_id = conversation["clientId"]
        conversation_signature = conversation["conversationSignature"]

        async with session.ws_connect(
                'wss://sydney.bing.com/sydney/ChatHub',
                autoping=False,
                headers=_HEADERS,
                proxy=proxy
        ) as wss:
            await wss.send_str(_format({'protocol': 'json', 'version': 1}))
            await wss.receive(timeout=900)
            await wss.send_str(_format({"type": 6}))

            struct = {
                'arguments': [
                    {
                        'optionsSets': getattr(_OptionSets, conversation_style.upper()).value,
                        'source': 'cib',
                        'allowedMessageTypes': _ALLOWED_MESSAGE_TYPES,
                        'sliceIds': _SLICE_IDS,
                        'traceId': os.urandom(16).hex(),
                        'isStartOfSession': True,
                        'message': {
                            "locale": locale,
                            "market": locale,
                            "region": locale[-2:],  # en-US -> US
                            "locationHints": _get_location_hint_from_locale(locale),
                            "author": "user",
                            "inputMethod": "Keyboard",
                            "text": prompt,
                            "messageType": random.choice(["Chat", "SearchQuery"]),
                            "imageUrl": image_url or None,
                        },
                        'conversationSignature': conversation_signature,
                        'participant': {
                            'id': client_id
                        },
                        'conversationId': conversation_id,
                        'previousMessages': [
                            {
                                "author": "user",
                                "description": context,
                                "contextType": "WebPage",
                                "messageType": "Context",
                            },
                        ]
                    }
                ],
                'invocationId': '0',
                'target': 'chat',
                'type': 4
            }

            await wss.send_str(_format(struct))

            retry_count = 5
            while True:
                if wss.closed:
                    break
                msg = await wss.receive(timeout=900)

                if not msg.data:
                    retry_count -= 1
                    if retry_count == 0:
                        raise Exception("No response from server")
                    continue

                if isinstance(msg.data, str):
                    objects = msg.data.split(_DELIMITER)
                else:
                    continue

                for obj in objects:
                    if int(time()) % 6 == 0:
                        await wss.send_str(_format({"type": 6}))
                    if not obj:
                        continue
                    response = json.loads(obj)
                    if response["type"] == 2:
                        if response["item"]["result"].get("error"):
                            raise Exception(
                                f"{response['item']['result']['value']}: {response['item']['result']['message']}")
                    yield response
                    if response["type"] == 2:
                        break


async def upload_image(filename, proxy):
    async with aiohttp.ClientSession(
            headers={"Referer": "https://www.bing.com/search?q=Bing+AI&showconv=1&FORM=hpcodx"}
    ) as session:
        url = "https://www.bing.com/images/kblob"

        payload = {
            "imageInfo": {},
            "knowledgeRequest": {
                "invokedSkills": ["ImageById"],
                "subscriptionId": "Bing.Chat.Multimodal",
                "invokedSkillsRequestData": {"enableFaceBlur": False},
                "convoData": {
                    "convoid": "",
                    "convotone": "Creative"
                }
            }
        }

        with open(filename, 'rb') as f:
            file_data = f.read()
            image_base64 = base64.b64encode(file_data)

        data = aiohttp.FormData()
        data.add_field('knowledgeRequest', json.dumps(payload), content_type="application/json")
        data.add_field('imageBase64', image_base64.decode('utf-8'), content_type="application/octet-stream")

        async with session.post(url, data=data, proxy=proxy) as resp:
            return (await resp.json())["blobId"]
