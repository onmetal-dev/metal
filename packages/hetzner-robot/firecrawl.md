[NAV ![](https://robot.hetzner.com/doc/webservice/images/navbar-cad8cdcb.png)](#)

[![](https://robot.hetzner.com/doc/webservice/images/logo-cb35228e.png)](https://www.hetzner.com "www.hetzner.com")

- [Preface](#preface)
  - [General](#general)
  - [Errors](#errors)
- [Server](#server)
  - [GET /server](#get-server)
  - [GET /server/{server-number}](#get-server-server-number)
  - [POST /server/{server-number}](#post-server-server-number)
  - [GET /server/{server-number}/cancellation](#get-server-server-number-cancellation)
  - [POST /server/{server-number}/cancellation](#post-server-server-number-cancellation)
  - [DELETE /server/{server-number}/cancellation](#delete-server-server-number-cancellation)
- [IP](#ip)
  - [GET /ip](#get-ip)
  - [GET /ip/{ip}](#get-ip-ip)
  - [POST /ip/{ip}](#post-ip-ip)
  - [GET /ip/{ip}/mac](#get-ip-ip-mac)
  - [PUT /ip/{ip}/mac](#put-ip-ip-mac)
  - [DELETE /ip/{ip}/mac](#delete-ip-ip-mac)
  - [GET /ip/{ip}/cancellation](#get-ip-ip-cancellation)
  - [POST /ip/{ip}/cancellation](#post-ip-ip-cancellation)
  - [DELETE /ip/{ip}/cancellation](#delete-ip-ip-cancellation)
- [Subnet](#subnet)
  - [GET /subnet](#get-subnet)
  - [GET /subnet/{net-ip}](#get-subnet-net-ip)
  - [POST /subnet/{net-ip}](#post-subnet-net-ip)
  - [GET /subnet/{net-ip}/mac](#get-subnet-net-ip-mac)
  - [PUT /subnet/{net-ip}/mac](#put-subnet-net-ip-mac)
  - [DELETE /subnet/{net-ip}/mac](#delete-subnet-net-ip-mac)
  - [GET /subnet/{net-ip}/cancellation](#get-subnet-net-ip-cancellation)
  - [POST /subnet/{net-ip}/cancellation](#post-subnet-net-ip-cancellation)
  - [DELETE /subnet/{ip}/cancellation](#delete-subnet-ip-cancellation)
- [Reset](#reset)
  - [GET /reset](#get-reset)
  - [GET /reset/{server-number}](#get-reset-server-number)
  - [POST /reset/{server-number}](#post-reset-server-number)
- [Failover](#failover)
  - [GET /failover](#get-failover)
  - [GET /failover/{failover-ip}](#get-failover-failover-ip)
  - [POST /failover/{failover-ip}](#post-failover-failover-ip)
  - [DELETE /failover/{failover-ip}](#delete-failover-failover-ip)
- [Wake on LAN](#wake-on-lan)
  - [GET /wol/{server-number}](#get-wol-server-number)
  - [POST /wol/{server-number}](#post-wol-server-number)
- [Boot configuration](#boot-configuration)
  - [GET /boot/{server-number}](#get-boot-server-number)
  - [GET /boot/{server-number}/rescue](#get-boot-server-number-rescue)
  - [POST /boot/{server-number}/rescue](#post-boot-server-number-rescue)
  - [DELETE /boot/{server-number}/rescue](#delete-boot-server-number-rescue)
  - [GET /boot/{server-number}/rescue/last](#get-boot-server-number-rescue-last)
  - [GET /boot/{server-number}/linux](#get-boot-server-number-linux)
  - [POST /boot/{server-number}/linux](#post-boot-server-number-linux)
  - [DELETE /boot/{server-number}/linux](#delete-boot-server-number-linux)
  - [GET /boot/{server-number}/linux/last](#get-boot-server-number-linux-last)
  - [GET /boot/{server-number}/vnc](#get-boot-server-number-vnc)
  - [POST /boot/{server-number}/vnc](#post-boot-server-number-vnc)
  - [DELETE /boot/{server-number}/vnc](#delete-boot-server-number-vnc)
  - [GET /boot/{server-number}/windows](#get-boot-server-number-windows)
  - [POST /boot/{server-number}/windows](#post-boot-server-number-windows)
  - [DELETE /boot/{server-number}/windows](#delete-boot-server-number-windows)
  - [GET /boot/{server-number}/plesk](#get-boot-server-number-plesk)
  - [POST /boot/{server-number}/plesk](#post-boot-server-number-plesk)
  - [DELETE /boot/{server-number}/plesk](#delete-boot-server-number-plesk)
  - [GET /boot/{server-number}/cpanel](#get-boot-server-number-cpanel)
  - [POST /boot/{server-number}/cpanel](#post-boot-server-number-cpanel)
  - [DELETE /boot/{server-number}/cpanel](#delete-boot-server-number-cpanel)
- [Reverse DNS](#reverse-dns)
  - [GET /rdns](#get-rdns)
  - [GET /rdns/{ip}](#get-rdns-ip)
  - [PUT /rdns/{ip}](#put-rdns-ip)
  - [POST /rdns/{ip}](#post-rdns-ip)
  - [DELETE /rdns/{ip}](#delete-rdns-ip)
- [Traffic](#traffic)
  - [POST /traffic](#post-traffic)
- [SSH keys](#ssh-keys)
  - [GET /key](#get-key)
  - [POST /key](#post-key)
  - [GET /key/{fingerprint}](#get-key-fingerprint)
  - [POST /key/{fingerprint}](#post-key-fingerprint)
  - [DELETE /key/{fingerprint}](#delete-key-fingerprint)
- [Server ordering](#server-ordering)
  - [Activation](#activation)
  - [Notes](#notes)
  - [GET /order/server/product](#get-order-server-product)
  - [GET /order/server/product/{product-id}](#get-order-server-product-product-id)
  - [GET /order/server/transaction](#get-order-server-transaction)
  - [POST /order/server/transaction](#post-order-server-transaction)
  - [GET /order/server/transaction/{id}](#get-order-server-transaction-id)
  - [GET /order/server_market/product](#get-order-server_market-product)
  - [GET /order/server_market/product/{product-id}](#get-order-server_market-product-product-id)
  - [GET /order/server_market/transaction](#get-order-server_market-transaction)
  - [POST /order/server_market/transaction](#post-order-server_market-transaction)
  - [GET /order/server_market/transaction/{id}](#get-order-server_market-transaction-id)
  - [GET /order/server_addon/{server-number}/product](#get-order-server_addon-server-number-product)
  - [GET /order/server_addon/transaction](#get-order-server_addon-transaction)
  - [POST /order/server_addon/transaction](#post-order-server_addon-transaction)
  - [GET /order/server_addon/transaction/{id}](#get-order-server_addon-transaction-id)
- [Storage Box](#storage-box)
  - [GET /storagebox](#get-storagebox)
  - [GET /storagebox/{storagebox-id}](#get-storagebox-storagebox-id)
  - [POST /storagebox/{storagebox-id}](#post-storagebox-storagebox-id)
  - [POST /storagebox/{storagebox-id}/password](#post-storagebox-storagebox-id-password)
  - [GET /storagebox/{storagebox-id}/snapshot](#get-storagebox-storagebox-id-snapshot)
  - [POST /storagebox/{storagebox-id}/snapshot](#post-storagebox-storagebox-id-snapshot)
  - [DELETE /storagebox/{storagebox-id}/snapshot/{snapshot-name}](#delete-storagebox-storagebox-id-snapshot-snapshot-name)
  - [POST /storagebox/{storagebox-id}/snapshot/{snapshot-name}](#post-storagebox-storagebox-id-snapshot-snapshot-name)
  - [POST /storagebox/{storagebox-id}/snapshot/{snapshot-name}/comment](#post-storagebox-storagebox-id-snapshot-snapshot-name-comment)
  - [GET /storagebox/{storagebox-id}/snapshotplan](#get-storagebox-storagebox-id-snapshotplan)
  - [POST /storagebox/{storagebox-id}/snapshotplan](#post-storagebox-storagebox-id-snapshotplan)
  - [GET /storagebox/{storagebox-id}/subaccount](#get-storagebox-storagebox-id-subaccount)
  - [POST /storagebox/{storagebox-id}/subaccount](#post-storagebox-storagebox-id-subaccount)
  - [PUT /storagebox/{storagebox-id}/subaccount/{sub-account-username}](#put-storagebox-storagebox-id-subaccount-sub-account-username)
  - [DELETE /storagebox/{storagebox-id}/subaccount/{sub-account-username}](#delete-storagebox-storagebox-id-subaccount-sub-account-username)
  - [POST /storagebox/{storagebox-id}/subaccount/{sub-account-username}/password](#post-storagebox-storagebox-id-subaccount-sub-account-username-password)
- [Firewall](#firewall)
  - [GET /firewall/{server-id}](#get-firewall-server-id)
  - [POST /firewall/{server-id}](#post-firewall-server-id)
  - [DELETE /firewall/{server-id}](#delete-firewall-server-id)
  - [GET /firewall/template](#get-firewall-template)
  - [POST /firewall/template](#post-firewall-template)
  - [GET /firewall/template/{template-id}](#get-firewall-template-template-id)
  - [POST /firewall/template/{template-id}](#post-firewall-template-template-id)
  - [DELETE /firewall/template/{template-id}](#delete-firewall-template-template-id)
- [vSwitch](#vswitch)
  - [GET /vswitch](#get-vswitch)
  - [POST /vswitch](#post-vswitch)
  - [GET /vswitch/{vswitch-id}](#get-vswitch-vswitch-id)
  - [POST /vswitch/{vswitch-id}](#post-vswitch-vswitch-id)
  - [DELETE /vswitch/{vswitch-id}](#delete-vswitch-vswitch-id)
  - [POST /vswitch/{vswitch-id}/server](#post-vswitch-vswitch-id-server)
  - [DELETE /vswitch/{vswitch-id}/server](#delete-vswitch-vswitch-id-server)
- [PHP Client](#php-client)

- [Documentation powered by Slate](https://github.com/slatedocs/slate)
- [![English](https://robot.hetzner.com/doc/webservice/images/gb-a98ec259.png%20%22English%22) English](en.html)
  [![Deutsch](https://robot.hetzner.com/doc/webservice/images/de-1daf2d67.png%20%22Deutsch%22) Deutsch](de.html)

- [![Login Robot](https://robot.hetzner.com/doc/webservice/images/favicon-077bf19b.ico%20%22Login%20Robot%22) Login Robot](https://robot.hetzner.com)

- [Legal](https://www.hetzner.com/rechtliches/impressum)
  | [Data privacy](https://www.hetzner.com/rechtliches/datenschutz)

# Preface

## General

The interface is based on the HTTP protocol; therefore, you can use any HTTP library with it. You can use a simple command-line client like [curl](https://curl.haxx.se/)
.

To be able to use the interface, a web service user is required. You can create this user in Robot via the user menu in the upper right corner under "Settings" -> "Web service and app settings".

- POST parameters are transferred in the format "application/x-www-form-urlencoded".
- The response format is [JSON](https://json.org)
  ; by appending ".yaml", the response format is set to [YAML](http://yaml.org/spec/1.1/)
  .
- If the query was successful, the HTTP status code 200 OK is returned. If a new resource was created, the HTTP status code is set to 201 CREATED.
- If there is an error, the appropriate HTTP error code is set.
- Authentication is done via HTTP Basic Auth.
- The webservice is accessible only via HTTPS.
- Domain registration is not available via the Robot Webservice, but there is a mail interface for the Domain Registration Robot. You can find more information on [Hetzner Docs](https://docs.hetzner.com/robot/domain-registration-robot/faq#does-the-domain-registration-robot-have-an-interfaceapi)
  .

### URL

https://robot-ws.your-server.de

## Errors

### Error format

    {
      "error":
      {
        "status": 404,
        "code": "BOOT_NOT_AVAILABLE",
        "message": "No boot configuration available for this server"
      }
    }

error (Object)

|                  |                        |
| ---------------- | ---------------------- |
| status (Integer) | HTTP Status Code       |
| code (String)    | Specific error code    |
| message (String) | Specific error message |

### Error format invalid input

error (Object)

|                  |                                           |
| ---------------- | ----------------------------------------- |
| status (Integer) | 400                                       |
| code (String)    | INVALID_INPUT                             |
| message (String) | invalid input                             |
| missing (Array)  | Array of missing input parameters or null |
| invalid (Array)  | Array of invalid input paramaters or null |

### Authentication error

If authentication fails, the HTTP status "401 - Unauthorized" is returned. Please note that the IP from which you attempt to access will be blocked for 10 minutes after 3 failed login attempts.

### Request limit

If the request limit is reached, the HTTP status "403 - Forbidden" is returned.

### Error format request limit

error (Object)

|                       |                          |
| --------------------- | ------------------------ |
| status (Integer)      | 403                      |
| code (String)         | RATE_LIMIT_EXCEEDED      |
| max_request (Integer) | Maximum allowed requests |
| interval (Integer)    | Time interval in seconds |
| message (String)      | rate limit exceeded      |

### Unavailability due to maintenance

If the webservice is unavailable due to maintenance, the HTTP Status "503 - Service Unavailable" is returned.

# Server

## GET /server

    curl -u "user:password" https://robot-ws.your-server.de/server


    [\
      {\
        "server":{\
          "server_ip":"123.123.123.123",\
          "server_ipv6_net":"2a01:f48:111:4221::",\
          "server_number":321,\
          "server_name":"server1",\
          "product":"DS 3000",\
          "dc":"NBG1-DC1",\
          "traffic":"5 TB",\
          "status":"ready",\
          "cancelled":false,\
          "paid_until":"2010-09-02",\
          "ip":[\
            "123.123.123.123"\
          ],\
          "subnet":[\
            {\
              "ip":"2a01:4f8:111:4221::",\
              "mask":"64"\
            }\
          ]\
        }\
      },\
      {\
        "server":{\
          "server_ip":"123.123.123.124",\
          "server_ipv6_net":"2a01:f48:111:4221::",\
          "server_number":421,\
          "server_name":"server2",\
          "product":"X5",\
          "dc":"FSN1-DC10",\
          "traffic":"2 TB",\
          "status":"ready",\
          "cancelled":false,\
          "paid_until":"2010-06-11",\
          "ip":[\
            "123.123.123.124"\
          ],\
          "subnet":null\
        }\
      }\
    ]

### Description

Query data of all servers

### Request limit

200 requests per 1 hour

### Output

(Array)server (Object)

|                          |                                                              |
| ------------------------ | ------------------------------------------------------------ |
| server_ip (String)       | Server main IP address                                       |
| server_ipv6_net (String) | Server main IPv6 net address                                 |
| server_number (Integer)  | Server ID                                                    |
| server_name (String)     | Server name                                                  |
| product (String)         | Server product name                                          |
| dc (String)              | Data center                                                  |
| traffic (String)         | Free traffic quota, 'unlimited' in case of unlimited traffic |
| status (String)          | Server status ("ready" or "in process")                      |
| cancelled (Boolean)      | Status of server cancellation                                |
| paid_until (String)      | Paid until date                                              |
| ip (Array)               | Array of assigned single IP addresses                        |
| subnet (Array)           | Array of assigned subnets                                    |

### Errors

| Status | Code             | Description     |
| ------ | ---------------- | --------------- |
| 404    | SERVER_NOT_FOUND | No server found |

## GET /server/{server-number}

    curl -u "user:password" https://robot-ws.your-server.de/server/321


    {
      "server":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:f48:111:4221::",
        "server_number":321,
        "server_name":"server1",
        "product":"EQ 8",
        "dc":"NBG1-DC1",
        "traffic":"5 TB",
        "status":"ready",
        "cancelled":false,
        "paid_until":"2010-08-04",
        "ip":[\
          "123.123.123.123"\
        ],
        "subnet":[\
          {\
            "ip":"2a01:4f8:111:4221::",\
            "mask":"64"\
          }\
        ],
        "reset":true,
        "rescue":true,
        "vnc":true,
        "windows":true,
        "plesk":true,
        "cpanel":true,
        "wol":true,
        "hot_swap":true,
        "linked_storagebox":12345
      }
    }

### Description

Query server data for a specific server

### Request limit

200 requests per 1 hour

### Output

server (Object)

|                             |                                                              |
| --------------------------- | ------------------------------------------------------------ |
| server_ip (String)          | Server main IP address                                       |
| server_ipv6_net (String)    | Server main IPv6 net address                                 |
| server_number (Integer)     | Server ID                                                    |
| server_name (String)        | Server name                                                  |
| product (String)            | Server product name                                          |
| dc (String)                 | Data center                                                  |
| traffic (String)            | Free traffic quota, 'unlimited' in case of unlimited traffic |
| status (String)             | Server status ("ready" or "in process")                      |
| cancelled (Boolean)         | Status of server cancellation                                |
| paid_until (String)         | Paid until date                                              |
| ip (Array)                  | Array of assigned single IP addresses                        |
| subnet (Array)              | Array of assigned subnets                                    |
| reset (Boolean)             | Flag of reset system availability                            |
| rescue (Boolean)            | Flag of Rescue System availability                           |
| vnc (Boolean)               | Flag of VNC installation availability                        |
| windows (Boolean)           | Flag of Windows installation availability                    |
| plesk (Boolean)             | Flag of Plesk installation availability                      |
| cpanel (Boolean)            | Flag of cPanel installation availability                     |
| wol (Boolean)               | Flag of Wake On Lan availability                             |
| hot_swap (Boolean)          | Flag of Hot Swap availability                                |
| linked_storagebox (Integer) | Linked Storage Box ID                                        |

### Errors

| Status | Code             | Description                              |
| ------ | ---------------- | ---------------------------------------- |
| 404    | SERVER_NOT_FOUND | Server with id {server-number} not found |

### Deprecations

|                                     |                                                                        |
| ----------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /server/{server-ip} | The main IPv4 address may be used alternatively to specify the server. |

## POST /server/{server-number}

    curl -u "user:password" https://robot-ws.your-server.de/server/321 -d server_name=server1


    {
      "server":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "server_name":"server1",
        "product":"EQ 8",
        "dc":"NBG1-DC1",
        "traffic":"5 TB",
        "status":"ready",
        "cancelled":false,
        "paid_until":"2010-08-04",
        "ip":[\
          "123.123.123.123"\
        ],
        "subnet":[\
          {\
            "ip":"2a01:4f8:111:4221::",\
            "mask":"64"\
          }\
        ],
        "reset":true,
        "rescue":true,
        "vnc":true,
        "windows":true,
        "plesk":true,
        "cpanel":true,
        "wol":true,
        "hot_swap":true
      }
    }

### Description

Update server name for a specific server

### Request limit

200 requests per 1 hour

### Input

| Name        | Description |
| ----------- | ----------- |
| server_name | Server name |

### Output

server (Object)

|                             |                                                              |
| --------------------------- | ------------------------------------------------------------ |
| server_ip (String)          | Server main IP address                                       |
| server_ipv6_net (String)    | Server main IPv6 net address                                 |
| server_number (Integer)     | Server ID                                                    |
| server_name (String)        | Server name                                                  |
| product (String)            | Server product name                                          |
| dc (String)                 | Data center                                                  |
| traffic (String)            | Free traffic quota, 'unlimited' in case of unlimited traffic |
| status (String)             | Server status ("ready" or "in process")                      |
| cancelled (Boolean)         | Status of server cancellation                                |
| paid_until (String)         | Paid until date                                              |
| ip (Array)                  | Array of assigned single IP addresses                        |
| subnet (Array)              | Array of assigned subnets                                    |
| reset (Boolean)             | Flag of reset system availability                            |
| rescue (Boolean)            | Flag of Rescue System availability                           |
| vnc (Boolean)               | Flag of VNC installation availability                        |
| windows (Boolean)           | Flag of Windows installation availability                    |
| plesk (Boolean)             | Flag of Plesk installation availability                      |
| cpanel (Boolean)            | Flag of cPanel installation availability                     |
| wol (Boolean)               | Flag of Wake On Lan availability                             |
| hot_swap (Boolean)          | Flag of Hot Swap availability                                |
| linked_storagebox (Integer) | Linked Storage Box ID                                        |

### Errors

| Status | Code             | Description                              |
| ------ | ---------------- | ---------------------------------------- |
| 400    | INVALID_INPUT    | Invalid input parameters                 |
| 404    | SERVER_NOT_FOUND | Server with id {server-number} not found |

### Deprecations

|                                      |                                                                        |
| ------------------------------------ | ---------------------------------------------------------------------- |
| @deprecated POST /server/{server-ip} | The main IPv4 address may be used alternatively to specify the server. |

## GET /server/{server-number}/cancellation

    curl -u "user:password" https://robot-ws.your-server.de/server/321/cancellation


    {
      "cancellation":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "server_name":"server1",
        "earliest_cancellation_date":"2014-04-15",
        "cancelled":false,
        "reservation_possible":false,
        "reserved":false,
        "cancellation_date":null,
        "cancellation_reason":[\
          "Upgrade to a new server",\
          "Dissatisfied with the hardware",\
          "Dissatisfied with the support",\
          "Dissatisfied with the network",\
          "Dissatisfied with the IP\/subnet allocation",\
          "Dissatisfied with the Robot webinterface",\
          "Dissatisfied with the official Terms and Conditions",\
          "Server no longer necessary due to project ending",\
          "Server too expensive"\
        ]
      }
    }

### Description

Query cancellation data for a server

### Request limit

200 requests per 1 hour

### Output

cancellation (Object)

|                                     |                                                                                                     |
| ----------------------------------- | --------------------------------------------------------------------------------------------------- |
| server_ip (String)                  | Server main IP address                                                                              |
| server_ipv6_net (String)            | Server main IPv6 net address                                                                        |
| server_number (Integer)             | Server ID                                                                                           |
| server_name (String)                | Server name                                                                                         |
| earliest_cancellation_date (String) | Earliest possible cancellation date, format yyyy-MM-dd                                              |
| cancelled (Boolean)                 | Status of server cancellation                                                                       |
| reservation_possible (Boolean)      | Indicates whether the current server location is eligible for reservation after server cancellation |
| reservation (Boolean)               | Indicates whether the current server location will be reserved after server cancellation            |
| cancellation_date (String)          | Cancellation date if cancellation is active, format yyyy-MM-dd, otherwise null                      |
| cancellation_reason (Array\|String) | Array of possible cancellation reasons or cancellation reason if cancellation is active             |

### Errors

| Status | Code             | Description                              |
| ------ | ---------------- | ---------------------------------------- |
| 404    | SERVER_NOT_FOUND | Server with id {server-number} not found |

### Deprecations

|                                                  |                                                                        |
| ------------------------------------------------ | ---------------------------------------------------------------------- |
| @deprecated GET /server/{server-ip}/cancellation | The main IPv4 address may be used alternatively to specify the server. |

## POST /server/{server-number}/cancellation

    curl -u "user:password" https://robot-ws.your-server.de/server/321/cancellation -d 'cancellation_date=2014-04-15'


    {
      "cancellation":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "server_name":"server1",
        "earliest_cancellation_date":"2014-04-15",
        "cancelled":true,
        "reserved":false,
        "reservation_possible":false,
        "cancellation_date":"2014-04-15",
        "cancellation_reason":null
      }
    }

### Description

Cancel a server

### Request limit

200 requests per 1 hour

### Input

| Name                | Description                                                                 |
| ------------------- | --------------------------------------------------------------------------- |
| cancellation_date   | Date to which the server should be cancelled or "now" to cancel immediately |
| cancellation_reason | Cancellation reason, optional                                               |
| reserve_location    | Whether server location shall be reserved ('true' or 'false')               |

Please note following regarding the reserve_location parameter:

- The parameter is mandatory if it is possible to reserve the server location. In this case the call to GET /server/{ip}/cancellation returns a response with reservation_possible set to 'true'.
- Otherwise, the parameter reserve_location is optional. However, if you enter an reserve_location parameter, the value for reserve_location must be 'false'.

### Output

cancellation (Object)

|                                     |                                                                                                     |
| ----------------------------------- | --------------------------------------------------------------------------------------------------- |
| server_ip (String)                  | Server main IP address                                                                              |
| server_ipv6_net (String)            | Server main IPv6 net address                                                                        |
| server_number (Integer)             | Server ID                                                                                           |
| server_name (String)                | Server name                                                                                         |
| earliest_cancellation_date (String) | Earliest possible cancellation date, format yyyy-MM-dd                                              |
| cancelled (Boolean)                 | Status of server cancellation                                                                       |
| reserved (Boolean)                  | Indicates whether the current server location will be reserved after server cancellation            |
| reservation_possible (Boolean)      | Indicates whether the current server location is eligible for reservation after server cancellation |
| cancellation_date (String)          | Cancellation date, format yyyy-MM-dd                                                                |
| cancellation_reason (String)        | Cancellation reason or null                                                                         |

### Errors

| Status | Code                                            | Description                                                                                           |
| ------ | ----------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| 404    | SERVER_NOT_FOUND                                | Server with id {server-number} not found                                                              |
| 409    | CONFLICT                                        | The server is already cancelled                                                                       |
| 409    | CONFLICT                                        | Cancellation of server is not possible because of active transfer                                     |
| 409    | SERVER_CANCELLATION_RESERVE_LOCATION_FALSE_ONLY | It is not possible to reserve the location. Remove parameter reserve_location or set value to 'false' |
| 500    | INTERNAL_ERROR                                  | Cancellation failed due to an internal error                                                          |

### Deprecations

|                                                   |                                                                        |
| ------------------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated POST /server/{server-ip}/cancellation | The main IPv4 address may be used alternatively to specify the server. |

## DELETE /server/{server-number}/cancellation

    curl -u "user:password" https://robot-ws.your-server.de/server/321/cancellation -X DELETE

### Description

Withdraw a server cancellation

### Request limit

200 requests per 1 hour

### Output

No output

### Errors

| Status | Code             | Description                                             |
| ------ | ---------------- | ------------------------------------------------------- |
| 404    | SERVER_NOT_FOUND | Server with id {server-number} not found                |
| 409    | CONFLICT         | The cancellation cannot be revoked                      |
| 500    | INTERNAL_ERROR   | Cancellation revocation failed due to an internal error |

### Deprecations

|                                                     |                                                                        |
| --------------------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated DELETE /server/{server-ip}/cancellation | The main IPv4 address may be used alternatively to specify the server. |

# IP

## GET /ip

    curl -u "user:password" https://robot-ws.your-server.de/ip


    [\
      {\
        "ip":{\
          "ip":"123.123.123.123",\
          "server_ip":"123.123.123.123",\
          "server_number":321,\
          "locked":false,\
          "separate_mac":null,\
          "traffic_warnings":false,\
          "traffic_hourly":50,\
          "traffic_daily":50,\
          "traffic_monthly":8\
        }\
      },\
      {\
        "ip":{\
          "ip":"124.124.124.124",\
          "server_ip":"123.123.123.123",\
          "server_number":321,\
          "locked":false,\
          "separate_mac":null,\
          "traffic_warnings":false,\
          "traffic_hourly":200,\
          "traffic_daily":2000,\
          "traffic_monthly":20\
        }\
      }\
    ]

### Description

Query list of all single IP addresses

### Request limit

5000 requests per 1 hour

### Input (optional)

| Name      | Description                                                            |
| --------- | ---------------------------------------------------------------------- |
| server_ip | Server main IP address, show only IP addresses assigned to this server |

### Output

(Array)ip (Object)

|                            |                                       |
| -------------------------- | ------------------------------------- |
| ip (String)                | IP address                            |
| server_ip (String)         | Servers main IP address               |
| server_number (Integer)    | Server ID                             |
| locked (Boolean)           | Status of locking                     |
| separate_mac (String)      | Separate MAC address, if not set null |
| traffic_warnings (Boolean) | True if traffic warnings are enabled  |
| traffic_hourly (Integer)   | Hourly traffic limit in MB            |
| traffic_daily (Integer)    | Daily traffic limit in MB             |
| traffic_monthly (Integer)  | Monthly traffic limit in GB           |

### Errors

| Status | Code      | Description           |
| ------ | --------- | --------------------- |
| 404    | NOT_FOUND | No IP addresses found |

## GET /ip/{ip}

    curl -u "user:password" https://robot-ws.your-server.de/ip/123.123.123.123


    {
      "ip":{
        "ip":"123.123.123.123",
        "gateway":"123.123.123.97",
        "mask":27,
        "broadcast":"123.123.123.127",
        "server_ip":"123.123.123.123",
        "server_number":321,
        "locked":false,
        "separate_mac":null,
        "traffic_warnings":false,
        "traffic_hourly":50,
        "traffic_daily":50,
        "traffic_monthly":8
      }
    }

### Description

Query data for a specific IP address

### Request limit

5000 requests per 1 hour

### Output

ip (Object)

|                            |                                       |
| -------------------------- | ------------------------------------- |
| ip (String)                | IP address                            |
| gateway (String)           | Gateway                               |
| mask (Integer)             | Subnet mask in CIDR notation          |
| broadcast (String)         | Broadcast address                     |
| server_ip (String)         | Servers main IP address               |
| server_number (Integer)    | Server ID                             |
| locked (Boolean)           | Status of locking                     |
| separate_mac (String)      | Separate MAC address, if not set null |
| traffic_warnings (Boolean) | True if traffic warnings are enabled  |
| traffic_hourly (Integer)   | Hourly traffic limit in MB            |
| traffic_daily (Integer)    | Daily traffic limit in MB             |
| traffic_monthly (Integer)  | Monthly traffic limit in GB           |

### Errors

| Status | Code         | Description           |
| ------ | ------------ | --------------------- |
| 404    | IP_NOT_FOUND | No IP addresses found |

## POST /ip/{ip}

    curl -u "user:password" https://robot-ws.your-server.de/ip/123.123.123.123 -d traffic_warnings=true


    {
      "ip":{
        "ip":"123.123.123.123",
        "gateway":"123.123.123.97",
        "mask":27,
        "broadcast":"123.123.123.127",
        "server_ip":"123.123.123.123",
        "server_number":321,
        "locked":false,
        "separate_mac":null,
        "traffic_warnings":true,
        "traffic_hourly":50,
        "traffic_daily":50,
        "traffic_monthly":8
      }
    }

### Description

Update traffic warning options for an IP address

### Request limit

5000 requests per 1 hour

### Input

| Name             | Description                                  |
| ---------------- | -------------------------------------------- |
| traffic_warnings | Enable/disable traffic warnings (true,false) |
| traffic_hourly   | Hourly traffic limit in MB                   |
| traffic_daily    | Daily traffic limit in MB                    |
| traffic_monthly  | Monthly traffic limit in GB                  |

### Output

ip (Object)

|                            |                                       |
| -------------------------- | ------------------------------------- |
| ip (String)                | IP address                            |
| gateway (String)           | Gateway                               |
| mask (Integer)             | Subnet mask in CIDR notation          |
| broadcast (String)         | Broadcast address                     |
| server_ip (String)         | Servers main IP address               |
| server_number (Integer)    | Server ID                             |
| locked (Boolean)           | Status of locking                     |
| separate_mac (String)      | Separate MAC address, if not set null |
| traffic_warnings (Boolean) | True if traffic warnings are enabled  |
| traffic_hourly (Integer)   | Hourly traffic limit in MB            |
| traffic_daily (Integer)    | Daily traffic limit in MB             |
| traffic_monthly (Integer)  | Monthly traffic limit in GB           |

### Errors

| Status | Code                          | Description                                                      |
| ------ | ----------------------------- | ---------------------------------------------------------------- |
| 400    | INVALID_INPUT                 | Invalid input parameters                                         |
| 404    | IP_NOT_FOUND                  | No IP addresses found                                            |
| 500    | TRAFFIC_WARNING_UPDATE_FAILED | Updating traffic warning options failed due to an internal error |

## GET /ip/{ip}/mac

    curl -u "user:password" https://robot-ws.your-server.de/ip/123.123.123.123/mac


    {
      "mac":{
        "ip":"123.123.123.123",
        "mac":"00:21:85:62:3e:9c"
      }
    }

### Description

Query if it is possible to set a separate MAC address.Returns the MAC address if it is set.

### Request limit

5000 requests per 1 hour

### Output

mac (Object)

|              |             |
| ------------ | ----------- |
| ip (String)  | IP address  |
| mac (String) | MAC address |

### Errors

| Status | Code              | Description                                                          |
| ------ | ----------------- | -------------------------------------------------------------------- |
| 404    | IP_NOT_FOUND      | IP address not found                                                 |
| 404    | MAC_NOT_FOUND     | There is no separate MAC address set                                 |
| 404    | MAC_NOT_AVAILABLE | For this IP address it is not possible to set a separate MAC address |

## PUT /ip/{ip}/mac

    curl -u "user:password" https://robot-ws.your-server.de/ip/123.123.123.123/mac  -X PUT


    {
      "mac":{
        "ip":"123.123.123.123",
        "mac":"00:21:85:62:3e:9c"
      }
    }

### Description

Generate a separate MAC address

### Request limit

10 requests per 1 hour

### Input

No input

### Output

mac (Object)

|              |             |
| ------------ | ----------- |
| ip (String)  | IP address  |
| mac (String) | MAC address |

### Errors

| Status | Code              | Description                                                              |
| ------ | ----------------- | ------------------------------------------------------------------------ |
| 404    | IP_NOT_FOUND      | IP address not found                                                     |
| 404    | MAC_NOT_AVAILABLE | For this IP address it is not possible to set a separate MAC address     |
| 409    | MAC_ALREADY_SET   | There is already a separate MAC address set                              |
| 500    | MAC_FAILED        | The separate MAC address could not be generated due to an internal error |

## DELETE /ip/{ip}/mac

    curl -u "user:password" https://robot-ws.your-server.de/ip/123.123.123.123/mac -X DELETE


    {
      "mac":{
        "ip":"123.123.123.123",
        "mac":null
      }
    }

### Description

Remove a separate MAC address

### Request limit

10 requests per 1 hour

### Input

No input

### Output

mac (Object)

|              |            |
| ------------ | ---------- |
| ip (String)  | IP address |
| mac (String) | null       |

### Errors

| Status | Code              | Description                                                            |
| ------ | ----------------- | ---------------------------------------------------------------------- |
| 404    | IP_NOT_FOUND      | IP address not found                                                   |
| 404    | MAC_NOT_AVAILABLE | For this IP address it is not possible to set a separate MAC address   |
| 409    | MAC_NOT_FOUND     | There is no separate MAC address set                                   |
| 500    | MAC_FAILED        | The separate MAC address could not be removed due to an internal error |

## GET /ip/{ip}/cancellation

    curl -u "user:password" https://robot-ws.your-server.de/ip/123.123.123.123/cancellation


    {
      "cancellation":{
        "ip":"123.123.123.123",
        "server_number":321,
        "earliest_cancellation_date":"2022-02-11",
        "cancelled":false,
        "cancellation-date":null
      }
    }

### Description

Query cancellation data for an IP

### Request limit

200 requests per 1 hour

### Output

cancellation (Object)

|                                     |                                                                                |
| ----------------------------------- | ------------------------------------------------------------------------------ |
| ip (String)                         | IP address                                                                     |
| server_number (String)              | Server ID                                                                      |
| earliest_cancellation_date (String) | Earliest possible cancellation date, format yyyy-MM-dd                         |
| cancelled (Boolean)                 | This shows whether or not the IP has been earmarked for cancellation.          |
| cancellation_date (String)          | Cancellation date if cancellation is active, format yyyy-MM-dd, otherwise null |

### Errors

| Status | Code         | Description                          |
| ------ | ------------ | ------------------------------------ |
| 404    | IP_NOT_FOUND | IP address not found                 |
| 409    | CONFLICT     | It's not possible to cancel this IP. |

## POST /ip/{ip}/cancellation

    curl -u "user:password" https://robot-ws.your-server.de/ip/123.123.123.123/cancellation \
    --data-urlencode="cancellation_date=2022-02-11"


    {
      "cancellation":{
        "ip":"123.123.123.123",
        "server_number":321,
        "earliest_cancellation_date":"2022-02-11",
        "cancelled":true,
        "cancellation-date":"2022-02-11"
      }
    }

### Description

Cancel an IP address

### Request limit

200 requests per 1 hour

### Input

| Name              | Description                                                                              |
| ----------------- | ---------------------------------------------------------------------------------------- |
| cancellation_date | Date which you want the IP cancellation to go into effect or "now" to cancel immediately |

### Output

cancellation (Object)

|                                     |                                                                                |
| ----------------------------------- | ------------------------------------------------------------------------------ |
| ip (String)                         | IP address                                                                     |
| server_number (String)              | Server ID                                                                      |
| earliest_cancellation_date (String) | Earliest possible cancellation date, format yyyy-MM-dd                         |
| cancelled (Boolean)                 | This shows whether or not the IP has been earmarked for cancellation.          |
| cancellation_date (String)          | Cancellation date if cancellation is active, format yyyy-MM-dd, otherwise null |

### Errors

| Status | Code          | Description                                                                          |
| ------ | ------------- | ------------------------------------------------------------------------------------ |
| 400    | INVALID_INPUT | Invalid input parameters                                                             |
| 404    | IP_NOT_FOUND  | IP address not found                                                                 |
| 409    | CONFLICT      | The IP address cannot be cancelled due to the reason mentioned in the error message. |

## DELETE /ip/{ip}/cancellation

    curl -u "user:password" https://robot-ws.your-server.de/ip/123.123.123.123/cancellation -X DELETE


    {
      "cancellation":{
        "ip":"123.123.123.123",
        "server_number":321,
        "earliest_cancellation_date":"2022-02-11",
        "cancelled":false,
        "cancellation-date":null
      }
    }

### Description

Revoke an IP cancellation

### Request limit

200 requests per 1 hour

### Input

No input

### Output

cancellation (Object)

|                                     |                                                                                |
| ----------------------------------- | ------------------------------------------------------------------------------ |
| ip (String)                         | IP address                                                                     |
| server_number (String)              | Server ID                                                                      |
| earliest_cancellation_date (String) | Earliest possible cancellation date, format yyyy-MM-dd                         |
| cancelled (Boolean)                 | This shows whether or not the IP has been earmarked for cancellation.          |
| cancellation_date (String)          | Cancellation date if cancellation is active, format yyyy-MM-dd, otherwise null |

### Errors

| Status | Code         | Description                                                                                     |
| ------ | ------------ | ----------------------------------------------------------------------------------------------- |
| 404    | IP_NOT_FOUND | IP address not found                                                                            |
| 409    | CONFLICT     | The IP address cancellation cannot be revoked due to the reason mentioned in the error message. |

# Subnet

## GET /subnet

    curl -u "user:password" https://robot-ws.your-server.de/subnet


    [\
      {\
        "subnet":{\
          "ip":"123.123.123.123",\
          "mask":29,\
          "gateway":"123.123.123.123",\
          "server_ip":"88.198.123.123",\
          "server_number":321,\
          "failover":false,\
          "locked":false,\
          "traffic_warnings":false,\
          "traffic_hourly":100,\
          "traffic_daily":500,\
          "traffic_monthly":2\
        }\
      },\
      {\
        "subnet":{\
          "ip":"178.63.123.123",\
          "mask":25,\
          "gateway":"178.63.123.124",\
          "server_ip":null,\
          "server_number":421,\
          "failover":false,\
          "locked":false,\
          "traffic_warnings":false,\
          "traffic_hourly":100,\
          "traffic_daily":500,\
          "traffic_monthly":2\
        }\
      }\
    ]

### Description

Query list of all subnets

### Request limit

5000 requests per 1 hour

### Input (optional)

| Name      | Description                                                       |
| --------- | ----------------------------------------------------------------- |
| server_ip | Server main IP address, show only subnets assigned to this server |

### Output

(Array)subnet (Object)

|                            |                                      |
| -------------------------- | ------------------------------------ |
| ip (String)                | IP address                           |
| mask (Integer)             | Subnet mask in CIDR notation         |
| gateway (String)           | Subnet gateway                       |
| server_ip (String)         | Servers main IP address              |
| server_number (Integer)    | Server ID                            |
| failover (Boolean)         | True if subnet is a failover subnet  |
| locked (Boolean)           | Status of locking                    |
| traffic_warnings (Boolean) | True if traffic warnings are enabled |
| traffic_hourly (Integer)   | Hourly traffic limit in MB           |
| traffic_daily (Integer)    | Daily traffic limit in MB            |
| traffic_monthly (Integer)  | Monthly traffic limit in GB          |

### Errors

| Status | Code      | Description      |
| ------ | --------- | ---------------- |
| 404    | NOT_FOUND | No subnets found |

## GET /subnet/{net-ip}

    curl -u "user:password" https://robot-ws.your-server.de/subnet/123.123.123.123


    {
      "subnet":{
        "ip":"123.123.123.123",
        "mask":29,
        "gateway":"123.123.123.123",
        "server_ip":"88.198.123.123",
        "server_number":321,
        "failover":false,
        "locked":false,
        "traffic_warnings":false,
        "traffic_hourly":100,
        "traffic_daily":500,
        "traffic_monthly":2
      }
    }

### Description

Query data of a specific subnet

### Request limit

5000 requests per 1 hour

### Output

subnet (Object)

|                            |                                      |
| -------------------------- | ------------------------------------ |
| ip (String)                | IP address                           |
| mask (Integer)             | Subnet mask in CIDR notation         |
| gateway (String)           | Subnet gateway                       |
| server_ip (String)         | Servers main IP address              |
| server_number (Integer)    | Server ID                            |
| failover (Boolean)         | True if subnet is a failover subnet  |
| locked (Boolean)           | Status of locking                    |
| traffic_warnings (Boolean) | True if traffic warnings are enabled |
| traffic_hourly (Integer)   | Hourly traffic limit in MB           |
| traffic_daily (Integer)    | Daily traffic limit in MB            |
| traffic_monthly (Integer)  | Monthly traffic limit in GB          |

### Errors

| Status | Code             | Description      |
| ------ | ---------------- | ---------------- |
| 404    | SUBNET_NOT_FOUND | Subnet not found |

## POST /subnet/{net-ip}

    curl -u "user:password" https://robot-ws.your-server.de/subnet/123.123.123.123 -d traffic_warnings=true


    {
      "subnet":{
        "ip":"123.123.123.123",
        "mask":29,
        "gateway":"123.123.123.123",
        "server_ip":"88.198.123.123",
        "server_number":321,
        "failover":false,
        "locked":false,
        "traffic_warnings":true,
        "traffic_hourly":100,
        "traffic_daily":500,
        "traffic_monthly":2
      }
    }

### Description

Update traffic warning options for an subnet

### Request limit

5000 requests per 1 hour

### Input

| Name             | Description                                  |
| ---------------- | -------------------------------------------- |
| traffic_warnings | Enable/disable traffic warnings (true,false) |
| traffic_hourly   | Hourly traffic limit in MB                   |
| traffic_daily    | Daily traffic limit in MB                    |
| traffic_monthly  | Monthly traffic limit in GB                  |

### Output

subnet (Object)

|                            |                                      |
| -------------------------- | ------------------------------------ |
| ip (String)                | IP address                           |
| mask (Integer)             | Subnet mask in CIDR notation         |
| gateway (String)           | Subnet gateway                       |
| server_ip (String)         | Servers main IP address              |
| server_number (Integer)    | Server ID                            |
| failover (Boolean)         | True if subnet is a failover subnet  |
| locked (Boolean)           | Status of locking                    |
| traffic_warnings (Boolean) | True if traffic warnings are enabled |
| traffic_hourly (Integer)   | Hourly traffic limit in MB           |
| traffic_daily (Integer)    | Daily traffic limit in MB            |
| traffic_monthly (Integer)  | Monthly traffic limit in GB          |

### Errors

| Status | Code                          | Description                                                      |
| ------ | ----------------------------- | ---------------------------------------------------------------- |
| 400    | INVALID_INPUT                 | Invalid input parameters                                         |
| 404    | SUBNET_NOT_FOUND              | Subnet not found                                                 |
| 500    | TRAFFIC_WARNING_UPDATE_FAILED | Updating traffic warning options failed due to an internal error |

## GET /subnet/{net-ip}/mac

    curl -u "user:password" https://robot-ws.your-server.de/subnet/2a01:4f8:111:4221::/mac


    {
      "mac":{
        "ip":"2a01:4f8:111:4221::",
        "mask":"64",
        "mac":"00:21:85:62:3e:9c",
        "possible_mac":{
          "123.123.123.123":"00:21:85:62:3e:9c",
          "123.123.123.124":"00:21:85:62:3e:9d"
        }
      }
    }

### Description

Query if it is possible to set a separate MAC address.

### Request limit

5000 requests per 1 hour

### Output

mac (Object)

|                       |                              |
| --------------------- | ---------------------------- |
| ip (String)           | IP address                   |
| mask (String)         | Subnet mask in CIDR notation |
| mac (String)          | MAC address                  |
| possible_mac (Object) | Possible MAC addresses       |

### Errors

| Status | Code              | Description                                                          |
| ------ | ----------------- | -------------------------------------------------------------------- |
| 404    | SUBNET_NOT_FOUND  | Subnet not found                                                     |
| 404    | MAC_NOT_AVAILABLE | For this IP address it is not possible to set a separate MAC address |

## PUT /subnet/{net-ip}/mac

    curl -u "user:password" https://robot-ws.your-server.de/subnet/2a01:4f8:111:4221::/mac -X PUT -d 'mac=00:21:85:62:3e:9d'


    {
      "mac":{
        "ip":"2a01:4f8:111:4221::",
        "mask":"64",
        "mac":"00:21:85:62:3e:9d",
        "possible_mac":{
          "123.123.123.123":"00:21:85:62:3e:9c",
          "123.123.123.124":"00:21:85:62:3e:9d"
        }
      }
    }

### Description

Generate a separate MAC address

### Request limit

10 requests per 1 hour

### Input

| Name | Description        |
| ---- | ------------------ |
| mac  | Target MAC address |

### Output

mac (Object)

|                       |                              |
| --------------------- | ---------------------------- |
| ip (String)           | IP address                   |
| mask (String)         | Subnet mask in CIDR notation |
| mac (String)          | MAC address                  |
| possible_mac (Object) | Possible MAC addresses       |

### Errors

| Status | Code              | Description                                                              |
| ------ | ----------------- | ------------------------------------------------------------------------ |
| 404    | SUBNET_NOT_FOUND  | Subnet not found                                                         |
| 404    | MAC_NOT_AVAILABLE | For this IP address it is not possible to set a separate MAC address     |
| 500    | MAC_FAILED        | The separate MAC address could not be generated due to an internal error |

## DELETE /subnet/{net-ip}/mac

    curl -u "user:password" https://robot-ws.your-server.de/subnet/2a01:4f8:111:4221::/mac -X DELETE


    {
      "mac":{
        "ip":"2a01:4f8:111:4221::",
        "mask":"64",
        "mac":"00:21:85:62:3e:9c",
        "possible_mac":{
          "123.123.123.123":"00:21:85:62:3e:9c",
          "123.123.123.124":"00:21:85:62:3e:9d"
        }
      }
    }

### Description

Remove a separate MAC address and set it to the default value (The MAC address of the servers main IP address).### Request limit

10 requests per 1 hour

### Input

No input

### Output

mac (Object)

|                       |                              |
| --------------------- | ---------------------------- |
| ip (String)           | IP address                   |
| mask (String)         | Subnet mask in CIDR notation |
| mac (String)          | MAC address                  |
| possible_mac (Object) | Possible MAC addresses       |

### Errors

| Status | Code              | Description                                                            |
| ------ | ----------------- | ---------------------------------------------------------------------- |
| 404    | SUBNET_NOT_FOUND  | Subnet not found                                                       |
| 404    | MAC_NOT_AVAILABLE | For this IP address it is not possible to set a separate MAC address   |
| 500    | MAC_FAILED        | The separate MAC address could not be removed due to an internal error |

## GET /subnet/{net-ip}/cancellation

    curl -u "user:password" https://robot-ws.your-server.de/subnet/123.123.123.123/cancellation


    {
      "cancellation":{
        "ip":"123.123.123.123",
        "mask":"29",
        "server_number":321,
        "earliest_cancellation_date":"2022-02-11",
        "cancelled":false,
        "cancellation-date":null
      }
    }

### Description

Query cancellation data for a subnet

### Request limit

200 requests per 1 hour

### Output

cancellation (Object)

|                                     |                                                                                |
| ----------------------------------- | ------------------------------------------------------------------------------ |
| ip (String)                         | IP address                                                                     |
| mask (String)                       | Subnet mask in CIDR notation                                                   |
| server_number (String)              | Server ID                                                                      |
| earliest_cancellation_date (String) | Earliest possible cancellation date, format yyyy-MM-dd                         |
| cancelled (Boolean)                 | This shows whether or not the subnet is earmarked for cancellation.            |
| cancellation_date (String)          | Cancellation date if cancellation is active, format yyyy-MM-dd, otherwise null |

### Errors

| Status | Code         | Description                               |
| ------ | ------------ | ----------------------------------------- |
| 404    | IP_NOT_FOUND | Subnet not found                          |
| 409    | CONFLICT     | It is not possible to cancel this subnet. |

## POST /subnet/{net-ip}/cancellation

    curl -u "user:password" https://robot-ws.your-server.de/subnet/123.123.123.123/cancellation \
    --data-urlencode="cancellation_date=2022-02-11"


    {
      "cancellation":{
        "ip":"123.123.123.123",
        "mask":"29",
        "server_number":321,
        "earliest_cancellation_date":"2022-02-11",
        "cancelled":true,
        "cancellation-date":"2022-02-11"
      }
    }

### Description

Cancel a subnet

### Request limit

200 requests per 1 hour

### Input

| Name              | Description                                                                                  |
| ----------------- | -------------------------------------------------------------------------------------------- |
| cancellation_date | Date which you want the subnet cancellation to go into effect or "now" to cancel immediately |

### Output

cancellation (Object)

|                                     |                                                                                |
| ----------------------------------- | ------------------------------------------------------------------------------ |
| ip (String)                         | IP address                                                                     |
| mask (String)                       | Subnet mask in CIDR notation                                                   |
| server_number (String)              | Server ID                                                                      |
| earliest_cancellation_date (String) | Earliest possible cancellation date, format yyyy-MM-dd                         |
| cancelled (Boolean)                 | This shows whether or not the subnet is earmarked for cancellation.            |
| cancellation_date (String)          | Cancellation date if cancellation is active, format yyyy-MM-dd, otherwise null |

### Errors

| Status | Code          | Description                                                                      |
| ------ | ------------- | -------------------------------------------------------------------------------- |
| 400    | INVALID_INPUT | Invalid input parameters                                                         |
| 404    | IP_NOT_FOUND  | Subnet not found                                                                 |
| 409    | CONFLICT      | The subnet cannot be cancelled due to the reason mentioned in the error message. |

## DELETE /subnet/{ip}/cancellation

    curl -u "user:password" https://robot-ws.your-server.de/ip/123.123.123.123/cancellation -X DELETE


    {
      "cancellation":{
        "ip":"123.123.123.123",
        "mask":"29",
        "server_number":321,
        "earliest_cancellation_date":"2022-02-11",
        "cancelled":false,
        "cancellation-date":null
      }
    }

### Description

Revoke a subnet cancellation

### Request limit

200 requests per 1 hour

### Input

No input

### Output

cancellation (Object)

|                                     |                                                                                |
| ----------------------------------- | ------------------------------------------------------------------------------ |
| ip (String)                         | IP address                                                                     |
| mask (String)                       | Subnet mask in CIDR notation                                                   |
| server_number (String)              | Server ID                                                                      |
| earliest_cancellation_date (String) | Earliest possible cancellation date, format yyyy-MM-dd                         |
| cancelled (Boolean)                 | This shows whether or not the subnet is earmarked for cancellation.            |
| cancellation_date (String)          | Cancellation date if cancellation is active, format yyyy-MM-dd, otherwise null |

### Errors

| Status | Code         | Description                                                                                 |
| ------ | ------------ | ------------------------------------------------------------------------------------------- |
| 404    | IP_NOT_FOUND | Subnet not found                                                                            |
| 409    | CONFLICT     | The subnet cancellation cannot be revoked due to the reason mentioned in the error message. |

# Reset

## GET /reset

    curl -u "user:password" https://robot-ws.your-server.de/reset


    [\
      {\
        "reset":{\
          "server_ip":"123.123.123.123",\
          "server_ipv6_net":"2a01:4f8:111:4221::",\
          "server_number":321,\
          "type":[\
            "sw",\
            "hw",\
            "man"\
          ]\
        }\
      },\
      {\
        "reset":{\
          "server_ip":"111.111.111.111",\
          "server_ipv6_net":"2a01:4f8:111:4221::",\
          "server_number":111,\
          "type":[\
            "power",\
            "power_long",\
            "hw",\
            "man"\
          ]\
        }\
      }\
    ]

### Description

Query reset options for all servers

### Request limit

500 requests per 1 hour

### Output

(Array)reset (Object)

|                          |                              |
| ------------------------ | ---------------------------- |
| server_ip (String)       | Server main IP address       |
| server_ipv6_net (String) | Server main IPv6 net address |
| server_number (Integer)  | Server ID                    |
| type (Array)             | Available reset options      |

### Errors

| Status | Code      | Description                        |
| ------ | --------- | ---------------------------------- |
| 404    | NOT_FOUND | No servers with reset option found |

## GET /reset/{server-number}

> Query a server on which software reboots can be performed via the web service

    curl -u "user:password" https://robot-ws.your-server.de/reset/321


    {
      "reset":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "type":[\
          "sw",\
          "hw",\
          "man"\
        ],
        "operating_status":"not supported"
      }
    }

> Query a server on which the operating status of the server can be queried via the web service or on which the power button can be operated via the web service

    curl -u "user:password" https://robot-ws.your-server.de/reset/111.111.111.111


    {
      "reset":{
        "server_ip":"111.111.111.111",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":111,
        "type":[\
          "power",\
          "power_long",\
          "hw",\
          "man"\
        ],
        "operating_status":"running"
      }
    }

### Description

Query reset options for a specific server

### Request limit

500 requests per 1 hour

### Output

reset (Object)

|                           |                                        |
| ------------------------- | -------------------------------------- |
| server_ip (String)        | Server main IP address                 |
| server_ipv6_net (String)  | Server main IPv6 net address           |
| server_number (Integer)   | Server ID                              |
| type (Array)              | Available reset options                |
| operating_status (String) | Current operating status of the server |

### Errors

| Status | Code                | Description                              |
| ------ | ------------------- | ---------------------------------------- |
| 404    | SERVER_NOT_FOUND    | Server with id {server-number} not found |
| 404    | RESET_NOT_AVAILABLE | The server has no reset option           |

### Deprecations

|                                    |                                                                        |
| ---------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /reset/{server-ip} | The main IPv4 address may be used alternatively to specify the server. |

## POST /reset/{server-number}

    curl -u "user:password" https://robot-ws.your-server.de/reset/321 -d type=hw


    {
      "reset":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "type":"hw"
      }
    }

### Description

Execute reset on specific server

### Request limit

50 requests per hour

### Input

| Name | Description           |
| ---- | --------------------- |
| type | Reset type to execute |

### Output

reset (Object)

|                          |                              |
| ------------------------ | ---------------------------- |
| server_ip (String)       | Server main IP address       |
| server_ipv6_net (String) | Server main IPv6 net address |
| server_number (Integer)  | Server ID                    |
| type (String)            | Executed reset option        |

### Errors

| Status | Code                | Description                               |
| ------ | ------------------- | ----------------------------------------- |
| 400    | INVALID_INPUT       | Invalid input parameters                  |
| 404    | SERVER_NOT_FOUND    | Server with id {server-number} not found  |
| 404    | RESET_NOT_AVAILABLE | The server has no reset option            |
| 409    | RESET_MANUAL_ACTIVE | There is already a running manual reset   |
| 500    | RESET_FAILED        | Resetting failed due to an internal error |

### Deprecations

|                                     |                                                                        |
| ----------------------------------- | ---------------------------------------------------------------------- |
| @deprecated POST /reset/{server-ip} | The main IPv4 address may be used alternatively to specify the server. |

# Failover

IMPORTANT: For the proper use of the failover IP, it must be configured (not ordered) on all servers it can be switched to, regardless of where it is currently routed to.

## GET /failover

    curl -u "user:password" https://robot-ws.your-server.de/failover


    [\
      {\
        "failover":{\
          "ip":"123.123.123.123",\
          "netmask":"255.255.255.255",\
          "server_ip":"78.46.1.93",\
          "server_ipv6_net":"2a01:4f8:d0a:2003::",\
          "server_number":321,\
          "active_server_ip":"78.46.1.93"\
        }\
      },\
      {\
        "failover":{\
          "ip":"2a01:4f8:fff1::",\
          "netmask":"ffff:ffff:ffff:ffff::",\
          "server_ip":"78.46.1.93",\
          "server_ipv6_net":"2a01:4f8:d0a:2003::",\
          "server_number":321,\
          "active_server_ip":"2a01:4f8:d0a:2003::"\
        }\
      }\
    ]

### Description

Query failover data for all servers

### Request limit

100 requests per 1 hour

### Output

(Array)failover (Object)

|                           |                                       |
| ------------------------- | ------------------------------------- |
| ip (String)               | Failover net address                  |
| netmask (String)          | Failover netmask                      |
| server_ip (String)        | Main IP of related server             |
| server_ipv6_net (String)  | Main IPv6 net of related server       |
| server_number (Integer)   | Server ID                             |
| active_server_ip (String) | Main IP of current destination server |

### Errors

| Status | Code      | Description                    |
| ------ | --------- | ------------------------------ |
| 404    | NOT_FOUND | No failover IP addresses found |

## GET /failover/{failover-ip}

> IPv4

    curl -u "user:password" https://robot-ws.your-server.de/failover/123.123.123.123


    {
      "failover":{
        "ip":"123.123.123.123",
        "netmask":"255.255.255.255",
        "server_ip":"78.46.1.93",
        "server_ipv6_net":"2a01:4f8:d0a:2003::",
        "server_number":321,
        "active_server_ip":"78.46.1.93"
      }
    }

> IPv6

    curl -u "user:password" https://robot-ws.your-server.de/failover/2a01:4f8:fff1::


    {
      "failover":{
        "ip":"2a01:4f8:fff1::",
        "netmask":"ffff:ffff:ffff:ffff::",
        "server_ip":"78.46.1.93",
        "server_ipv6_net":"2a01:4f8:d0a:2003::",
        "server_number":321,
        "active_server_ip":"2a01:4f8:d0a:2003::"
      }
    }

### Description

Query specific failover IP address data

### Request limit

100 requests per 1 hour

### Output

failover (Object)

|                           |                                       |
| ------------------------- | ------------------------------------- |
| ip (String)               | Failover net address                  |
| netmask (String)          | Failover netmask                      |
| server_ip (String)        | Main IP of related server             |
| server_ipv6_net (String)  | Main IPv6 net of related server       |
| server_number (Integer)   | Server ID                             |
| active_server_ip (String) | Main IP of current destination server |

### Errors

| Status | Code      | Description                   |
| ------ | --------- | ----------------------------- |
| 404    | NOT_FOUND | Failover IP address not found |

## POST /failover/{failover-ip}

> IPv4

    curl -u "user:password" https://robot-ws.your-server.de/failover/123.123.123.123 \
    -d active_server_ip=124.124.124.124


    {
      "failover":{
        "ip":"123.123.123.123",
        "netmask":"255.255.255.255",
        "server_ip":"78.46.1.93",
        "server_ipv6_net":"2a01:4f8:d0a:2003::",
        "server_number":321,
        "active_server_ip":"124.124.124.124"
      }
    }

> IPv6

    curl -u "user:password" https://robot-ws.your-server.de/failover/2a01:4f8:fff1:: \
    -d active_server_ip=2a01:4f8:0:5176::


    {
      "failover":{
        "ip":"2a01:4f8:fff1::",
        "netmask":"ffff:ffff:ffff:ffff::",
        "server_ip":"78.46.1.93",
        "server_ipv6_net":"2a01:4f8:d0a:2003::",
        "server_number":321,
        "active_server_ip":"2a01:4f8:0:5176::"
      }
    }

### Description

Switch routing of failover IP address to another server

### Request limit

50 requests per hour

### Input

| Name             | Description                                                              |
| ---------------- | ------------------------------------------------------------------------ |
| active_server_ip | Main IP address of the server where the failover IP should be routed to. |

In case of IPv6 subnets, "active_server_ip" must be set to the subnet address of the server's main IPv6 subnet.

### Output

failover (Object)

|                           |                                       |
| ------------------------- | ------------------------------------- |
| ip (String)               | Failover net address                  |
| netmask (String)          | Failover netmask                      |
| server_ip (String)        | Main IP of related server             |
| server_ipv6_net (String)  | Main IPv6 net of related server       |
| server_number (Integer)   | Server ID                             |
| active_server_ip (String) | Main IP of current destination server |

### Errors

| Status | Code                          | Description                                                                |
| ------ | ----------------------------- | -------------------------------------------------------------------------- |
| 400    | INVALID_INPUT                 | Invalid input parameters                                                   |
| 404    | NOT_FOUND                     | Failover IP address not found                                              |
| 404    | FAILOVER_NEW_SERVER_NOT_FOUND | Destination server not found                                               |
| 409    | FAILOVER_ALREADY_ROUTED       | The failover IP address is already routed to the selected server           |
| 409    | FAILOVER_LOCKED               | Switching the failover IP address is blocked due to another active request |
| 500    | FAILOVER_FAILED               | Due to an internal error switching of the failover IP address failed       |
| 500    | FAILOVER_NOT_COMPLETE         | Due to an internal error switching of the failover IP address failed       |

## DELETE /failover/{failover-ip}

> IPv4

    curl -u "user:password" -X DELETE https://robot-ws.your-server.de/failover/123.123.123.123


    {
      "failover":{
        "ip":"123.123.123.123",
        "netmask":"255.255.255.255",
        "server_ip":"78.46.1.93",
        "server_ipv6_net":"2a01:4f8:d0a:2003::",
        "server_number":321,
        "active_server_ip":null
      }
    }

> IPv6

    curl -u "user:password" -X DELETE https://robot-ws.your-server.de/failover/2a01:4f8:fff1::


    {
      "failover":{
        "ip":"2a01:4f8:fff1::",
        "netmask":"ffff:ffff:ffff:ffff::",
        "server_ip":"78.46.1.93",
        "server_ipv6_net":"2a01:4f8:d0a:2003::",
        "server_number":321,
        "active_server_ip":null
      }
    }

### Description

Delete the routing of a failover IP

### Request limit

50 requests per hour

### Output

failover (Object)

|                           |                                       |
| ------------------------- | ------------------------------------- |
| ip (String)               | Failover net address                  |
| netmask (String)          | Failover netmask                      |
| server_ip (String)        | Main IP of related server             |
| server_ipv6_net (String)  | Main IPv6 net of related server       |
| server_number (Integer)   | Server ID                             |
| active_server_ip (String) | Main IP of current destination server |

### Errors

| Status | Code                  | Description                                                               |
| ------ | --------------------- | ------------------------------------------------------------------------- |
| 404    | NOT_FOUND             | Failover IP address not found                                             |
| 409    | FAILOVER_LOCKED       | Deleting the failover IP routing is blocked due to another active request |
| 500    | FAILOVER_FAILED       | Due to an internal error deleting the failover IP routing failed          |
| 500    | FAILOVER_NOT_COMPLETE | Due to an internal error deleting the failover IP routing failed          |

# Wake on LAN

## GET /wol/{server-number}

    curl -u "user:password" https://robot-ws.your-server.de/wol/321


    {
      "wol":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321
      }
    }

### Description

Query Wake On LAN data

### Request limit

500 requests per 1 hour

### Output

wol (Object)

|                          |                              |
| ------------------------ | ---------------------------- |
| server_ip (String)       | Server main IP address       |
| server_ipv6_net (String) | Server main IPv6 net address |
| server_number (Integer)  | Server ID                    |

### Errors

| Status | Code              | Description                                 |
| ------ | ----------------- | ------------------------------------------- |
| 404    | SERVER_NOT_FOUND  | Server not found                            |
| 404    | WOL_NOT_AVAILABLE | Wake On LAN is not available on this server |

### Deprecations

|                                  |                                                                        |
| -------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /wol/{server-ip} | The main IPv4 address may be used alternatively to specify the server. |

## POST /wol/{server-number}

    curl -u "user:password" https://robot-ws.your-server.de/wol/321 -d ''


    {
      "wol":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321
      }
    }

### Description

Send Wake On LAN packet to server

### Request limit

10 requests per 1 hour

### Output

wol (Object)

|                          |                              |
| ------------------------ | ---------------------------- |
| server_ip (String)       | Server main IP address       |
| server_ipv6_net (String) | Server main IPv6 net address |
| server_number (Integer)  | Server ID                    |

### Errors

| Status | Code              | Description                                                |
| ------ | ----------------- | ---------------------------------------------------------- |
| 404    | SERVER_NOT_FOUND  | Server not found                                           |
| 404    | WOL_NOT_AVAILABLE | Wake On LAN is not available on this server                |
| 500    | WOL_FAILED        | Sending Wake On LAN packet failed due to an internal error |

### Deprecations

|                                   |                                                                        |
| --------------------------------- | ---------------------------------------------------------------------- |
| @deprecated POST /wol/{server-ip} | The main IPv4 address may be used alternatively to specify the server. |

# Boot configuration

## GET /boot/{server-number}

    curl -u "user:password" https://robot-ws.your-server.de/boot/321


    {
      "boot":{
        "rescue":{
          "server_ip":"123.123.123.123",
          "server_ipv6_net":"2a01:4f8:111:4221::",
          "server_number":321,
          "os":[\
            "linux",\
            "vkvm"\
          ],
          "@deprecated arch":[\
            64,\
            32\
          ],
          "active":false,
          "password":null,
          "authorized_key":[\
    \
          ],
          "host_key":[\
    \
          ]
        },
        "linux":{
          "server_ip":"123.123.123.123",
          "server_ipv6_net":"2a01:4f8:111:4221::",
          "server_number":321,
          "dist":[\
            "CentOS 5.5 minimal",\
            "Debian 7.8 minimal"\
          ],
          "@deprecated arch":[\
            64,\
            32\
          ],
          "lang":[\
            "en"\
          ],
          "active":false,
          "password":null,
          "authorized_key":[\
    \
          ],
          "host_key":[\
    \
          ]
        },
        "vnc":{
          "server_ip":"123.123.123.123",
          "server_ipv6_net":"2a01:4f8:111:4221::",
          "server_number":321,
          "dist":[\
            "centOS-5.0",\
            "Fedora-6",\
            "openSUSE-10.2"\
          ],
          "@deprecated arch":[\
            64,\
            32\
          ],
          "lang":[\
            "de_DE",\
            "en_US"\
          ],
          "active":false,
          "password":null
        },
        "windows":{
          "server_ip":"123.123.123.123",
          "server_ipv6_net":"2a01:4f8:111:4221::",
          "server_number":321,
          "dist":null,
          "lang":null,
          "active":false,
          "password":null
        },
        "plesk":{
          "server_ip":"123.123.123.123",
          "server_ipv6_net":"2a01:4f8:111:4221::",
          "server_number":321,
          "dist":[\
            "CentOS 5.4 minimal",\
            "Debian 7.8 minimal"\
          ],
          "@deprecated arch":[\
            64,\
            32\
          ],
          "lang":[\
            "en",\
            "de"\
          ],
          "active":false,
          "password":null,
          "hostname":null
        },
        "cpanel":{
          "server_ip":"123.123.123.123",
          "server_ipv6_net":"2a01:4f8:111:4221::",
          "server_number":321,
          "dist":[\
            "CentOS 5.6 + cPanel"\
          ],
          "@deprecated arch":[\
            64\
          ],
          "lang":[\
            "en"\
          ],
          "active":false,
          "password":null,
          "hostname":null
        }
      }
    }

### Description

Query the current boot configuration status for a server. There can be only one configuration active at any time for one server.

### Request limit

500 requests per 1 hour

### Output

boot (Object)rescue (Object)

|                                   |                                                                     |
| --------------------------------- | ------------------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                              |
| server_ipv6_net (String)          | Server main IPv6 net address                                        |
| server_number (Integer)           | Server ID                                                           |
| os (Array\|String)                | Array of available operating systems or the active operating system |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture         |
| active (Boolean)                  | Current Rescue System status                                        |
| password (String)                 | Current Rescue System root password or null                         |
| authorized_key (Array)            | Authorized public SSH keys                                          |
| host_key (Array)                  | Host keys                                                           |

linux (Object)

|                                   |                                                             |
| --------------------------------- | ----------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                      |
| server_ipv6_net (String)          | Server main IPv6 net address                                |
| server_number (Integer)           | Server ID                                                   |
| dist (Array\|String)              | Array of available distributions or the active distributon  |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture |
| lang (Array\|String)              | Array of available languages or the active language         |
| active (Boolean)                  | Current Linux installation status                           |
| password (String)                 | Current Linux installation password or null                 |
| authorized_key (Array)            | Authorized public SSH keys                                  |
| host_key (Array)                  | Host keys                                                   |

vnc (Object)

|                                   |                                                             |
| --------------------------------- | ----------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                      |
| server_ipv6_net (String)          | Server main IPv6 net address                                |
| server_number (Integer)           | Server ID                                                   |
| dist (Array\|String)              | Array of available distributions or the active distributon  |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture |
| lang (Array\|String)              | Array of available languages or the active language         |
| active (Boolean)                  | Current VNC installation status                             |
| password (String)                 | Current VNC installation password or null                   |

windows (Object)

|                                   |                                                             |
| --------------------------------- | ----------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                      |
| server_ipv6_net (String)          | Server main IPv6 net address                                |
| server_number (Integer)           | Server ID                                                   |
| dist (Array\|String)              | Array of available distributions or the active distributon  |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture |
| lang (Array\|String)              | Array of available languages or the active language         |
| active (Boolean)                  | Current Windows installation status                         |
| password (String)                 | Current Windows installation password or null               |

plesk (Object)

|                                   |                                                             |
| --------------------------------- | ----------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                      |
| server_ipv6_net (String)          | Server main IPv6 net address                                |
| server_number (Integer)           | Server ID                                                   |
| dist (Array\|String)              | Array of available distributions or the active distributon  |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture |
| lang (Array\|String)              | Array of available languages or the active language         |
| active (Boolean)                  | Current Plesk installation status                           |
| password (String)                 | Current Plesk installation password or null                 |
| hostname (String)                 | Current Plesk installation hostname or null                 |

cpanel (Object)

|                                   |                                                             |
| --------------------------------- | ----------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                      |
| server_ipv6_net (String)          | Server main IPv6 net address                                |
| server_number (Integer)           | Server ID                                                   |
| dist (Array\|String)              | Array of available distributions or the active distributon  |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture |
| lang (Array\|String)              | Array of available languages or the active language         |
| active (Boolean)                  | Current cPanel installation status                          |
| password (String)                 | Current cPanel installation password or null                |
| hostname (String)                 | Current cPanel installation hostname or null                |

### Errors

| Status | Code               | Description                                     |
| ------ | ------------------ | ----------------------------------------------- |
| 404    | SERVER_NOT_FOUND   | Server with id {server-number} not found        |
| 404    | BOOT_NOT_AVAILABLE | No boot configuration available for this server |

### Deprecations

|                                   |                                                                        |
| --------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /boot/{server-ip} | The main IPv4 address may be used alternatively to specify the server. |

## GET /boot/{server-number}/rescue

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/rescue


    {
      "rescue":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "os":[\
          "linux",\
          "vkvm"\
        ],
        "@deprecated arch":[\
          64,\
          32\
        ],
        "active":false,
        "password":null,
        "authorized_key":[\
    \
        ],
        "host_key":[\
    \
        ]
      }
    }

### Description

Query boot options for the Rescue System

### Request limit

500 requests per 1 hour

### Output

rescue (Object)

|                                   |                                                                     |
| --------------------------------- | ------------------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                              |
| server_ipv6_net (String)          | Server main IPv6 net address                                        |
| server_number (Integer)           | Server ID                                                           |
| os (Array\|String)                | Array of available operating systems or the active operating system |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture         |
| active (Boolean)                  | Current Rescue System status                                        |
| password (String)                 | Current Rescue System root password or null                         |
| authorized_key (Array)            | Authorized public SSH keys                                          |
| host_key (Array)                  | Host keys                                                           |

### Errors

| Status | Code               | Description                                     |
| ------ | ------------------ | ----------------------------------------------- |
| 404    | SERVER_NOT_FOUND   | Server with id {server-number} not found        |
| 404    | BOOT_NOT_AVAILABLE | No boot configuration available for this server |

### Deprecations

|                                          |                                                                        |
| ---------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /boot/{server-ip}/rescue | The main IPv4 address may be used alternatively to specify the server. |

## POST /boot/{server-number}/rescue

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/rescue -d 'os=linux'


    {
      "rescue":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "os":"linux",
        "@deprecated arch":32,
        "active":true,
        "password":"jEt0dtUvomlyOwRr",
        "authorized_key":[\
    \
        ],
        "host_key":[\
    \
        ]
      }
    }

### Description

Activate Rescue System

### Request limit

500 requests per 1 hour

### Input

| Name             | Description                                     |
| ---------------- | ----------------------------------------------- |
| os               | Operating System                                |
| @deprecated arch | Architecture (optional, default: 64)            |
| authorized_key   | One or more SSH key fingerprints (optional)     |
| keyboard         | Desired keyboard layout (optional, default: us) |

### Output

rescue (Object)

|                            |                              |
| -------------------------- | ---------------------------- |
| server_ip (String)         | Server main IP address       |
| server_ipv6_net (String)   | Server main IPv6 net address |
| server_number (Integer)    | Server ID                    |
| os (String)                | Operating system             |
| @deprecated arch (Integer) | Architecture                 |
| active (Boolean)           | true                         |
| password (String)          | Rescue System root password  |
| authorized_key (Array)     | Authorized public SSH keys   |
| host_key (Array)           | Host keys                    |

### Errors

| Status | Code                   | Description                                                     |
| ------ | ---------------------- | --------------------------------------------------------------- |
| 400    | INVALID_INPUT          | Invalid input parameters                                        |
| 404    | SERVER_NOT_FOUND       | Server with id {server-number} not found                        |
| 404    | BOOT_NOT_AVAILABLE     | No boot configuration available for this server                 |
| 500    | BOOT_ACTIVATION_FAILED | Activation of the Rescue System failed due to an internal error |

### Deprecations

|                                           |                                                                        |
| ----------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated POST /boot/{server-ip}/rescue | The main IPv4 address may be used alternatively to specify the server. |

## DELETE /boot/{server-number}/rescue

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/rescue -X DELETE


    {
      "rescue":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "os":[\
          "linux",\
          "vkvm"\
        ],
        "@deprecated arch":[\
          64,\
          32\
        ],
        "active":false,
        "password":null,
        "authorized_key":[\
    \
        ],
        "host_key":[\
    \
        ]
      }
    }

### Description

Deactivate Rescue System

### Request limit

500 requests per 1 hour

### Output

rescue (Object)

|                          |                                      |
| ------------------------ | ------------------------------------ |
| server_ip (String)       | Server main IP address               |
| server_ipv6_net (String) | Server main IPv6 net address         |
| server_number (Integer)  | Server ID                            |
| os (Array)               | Array of available operating systems |
| @deprecated arch (Array) | Array of available architectures     |
| active (Boolean)         | false                                |
| password (String)        | null                                 |
| authorized_key (Array)   | Authorized public SSH keys           |
| host_key (Array)         | Host keys                            |

### Errors

| Status | Code                     | Description                                                       |
| ------ | ------------------------ | ----------------------------------------------------------------- |
| 404    | SERVER_NOT_FOUND         | Server with id {server-number} not found                          |
| 404    | BOOT_NOT_AVAILABLE       | No boot configuration available for this server                   |
| 500    | BOOT_DEACTIVATION_FAILED | Deactivation of the Rescue System failed due to an internal error |

### Deprecations

|                                                     |                                                                        |
| --------------------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated POST /boot/{server-ip}/rescue -X DELETE | The main IPv4 address may be used alternatively to specify the server. |

## GET /boot/{server-number}/rescue/last

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/rescue/last


    {
      "rescue":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "os":"linux",
        "@deprecated arch":64,
        "active":false,
        "password":null,
        "authorized_key":[\
    \
        ],
        "host_key":[\
    \
        ]
      }
    }

### Description

Show data of last rescue activation

### Request limit

500 requests per 1 hour

### Output

rescue (Object)

|                            |                                             |
| -------------------------- | ------------------------------------------- |
| server_ip (String)         | Server main IP address                      |
| server_ipv6_net (String)   | Server main IPv6 net address                |
| server_number (Integer)    | Server ID                                   |
| os (String)                | Operating system                            |
| @deprecated arch (Integer) | Architecture                                |
| active (Boolean)           | Current Rescue System status                |
| password (String)          | Current Rescue System root password or null |
| authorized_key (Array)     | Authorized public SSH keys                  |
| host_key (Array)           | Host keys                                   |

### Errors

| Status | Code               | Description                                     |
| ------ | ------------------ | ----------------------------------------------- |
| 404    | SERVER_NOT_FOUND   | Server with id {server-number} not found        |
| 404    | BOOT_NOT_AVAILABLE | No boot configuration available for this server |

## GET /boot/{server-number}/linux

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/linux


    {
      "linux":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":[\
          "CentOS 5.5 minimal",\
          "Debian 7.8 minimal"\
        ],
        "@deprecated arch":[\
          64,\
          32\
        ],
        "lang":[\
          "en"\
        ],
        "active":false,
        "password":null,
        "authorized_key":[\
    \
        ],
        "host_key":[\
    \
        ]
      }
    }

### Description

Query boot options for the Linux installation

### Request limit

500 requests per 1 hour

### Output

linux (Object)

|                                   |                                                             |
| --------------------------------- | ----------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                      |
| server_ipv6_net (String)          | Server main IPv6 net address                                |
| server_number (Integer)           | Server ID                                                   |
| dist (Array\|String)              | Array of available distributions or the active distributon  |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture |
| lang (Array\|String)              | Array of available languages or the active language         |
| active (Boolean)                  | Current Linux installation status                           |
| password (String)                 | Current Linux installation password or null                 |
| authorized_key (Array)            | Authorized public SSH keys                                  |
| host_key (Array)                  | Host keys                                                   |

### Errors

| Status | Code               | Description                                     |
| ------ | ------------------ | ----------------------------------------------- |
| 404    | SERVER_NOT_FOUND   | Server with id {server-number} not found        |
| 404    | BOOT_NOT_AVAILABLE | No boot configuration available for this server |

### Deprecations

|                                         |                                                                        |
| --------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /boot/{server-ip}/linux | The main IPv4 address may be used alternatively to specify the server. |

## POST /boot/{server-number}/linux

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/linux -d 'dist=CentOS 5.5 minimal&lang=en'


    {
      "linux":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":"CentOS 5.5 minimal",
        "@deprecated arch":32,
        "lang":"en",
        "active":true,
        "password":"jEt0dtUvomlyOwRr",
        "authorized_key":[\
    \
        ],
        "host_key":[\
    \
        ]
      }
    }

### Description

Activate Linux installation

### Request limit

500 requests per 1 hour

### Input

| Name             | Description                                 |
| ---------------- | ------------------------------------------- |
| dist             | Distribution                                |
| @deprecated arch | Architecture (optional, default: 64)        |
| lang             | Language                                    |
| authorized_key   | One or more SSH key fingerprints (optional) |

### Output

linux (Object)

|                            |                              |
| -------------------------- | ---------------------------- |
| server_ip (String)         | Server main IP address       |
| server_ipv6_net (String)   | Server main IPv6 net address |
| server_number (Integer)    | Server ID                    |
| dist (String)              | Distribution                 |
| @deprecated arch (Integer) | Architecture                 |
| lang (String)              | Language                     |
| active (Boolean)           | true                         |
| password (String)          | Linux installation password  |
| authorized_key (Array)     | Authorized public SSH keys   |
| host_key (Array)           | Host keys                    |

### Errors

| Status | Code                   | Description                                                          |
| ------ | ---------------------- | -------------------------------------------------------------------- |
| 400    | INVALID_INPUT          | Invalid input parameters                                             |
| 404    | SERVER_NOT_FOUND       | Server with id {server-number} not found                             |
| 404    | BOOT_NOT_AVAILABLE     | No boot configuration available for this server                      |
| 500    | BOOT_ACTIVATION_FAILED | Activation of the Linux installation failed due to an internal error |

### Deprecations

|                                          |                                                                        |
| ---------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated POST /boot/{server-ip}/linux | The main IPv4 address may be used alternatively to specify the server. |

## DELETE /boot/{server-number}/linux

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/linux -X DELETE


    {
      "linux":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":[\
          "CentOS 5.5 minimal",\
          "Debian 7.8 minimal"\
        ],
        "@deprecated arch":[\
          64,\
          32\
        ],
        "lang":[\
          "en"\
        ],
        "active":false,
        "password":null,
        "authorized_key":[\
    \
        ],
        "host_key":[\
    \
        ]
      }
    }

### Description

Deactivate Linux installation

### Request limit

500 requests per 1 hour

### Output

linux (Object)

|                          |                                  |
| ------------------------ | -------------------------------- |
| server_ip (String)       | Server main IP address           |
| server_number (Integer)  | Server ID                        |
| dist (Array)             | Array of available distributions |
| @deprecated arch (Array) | Array of available architectures |
| lang (Array)             | Array of available languages     |
| active (Boolean)         | false                            |
| password (String)        | null                             |
| authorized_key (Array)   | Authorized public SSH keys       |
| host_key (Array)         | Host keys                        |

### Errors

| Status | Code                     | Description                                                            |
| ------ | ------------------------ | ---------------------------------------------------------------------- |
| 404    | SERVER_NOT_FOUND         | Server with id {server-number} not found                               |
| 404    | BOOT_NOT_AVAILABLE       | No boot configuration available for this server                        |
| 500    | BOOT_DEACTIVATION_FAILED | Deactivation of the Linux installation failed due to an internal error |

### Deprecations

|                                                    |                                                                        |
| -------------------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated POST /boot/{server-ip}/linux -X DELETE | The main IPv4 address may be used alternatively to specify the server. |

## GET /boot/{server-number}/linux/last

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/linux/last


    {
      "linux":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":"CentOS 5.5 minimal",
        "@deprecated arch":32,
        "lang":"en",
        "active":true,
        "password":"jEt0dtUvomlyOwRr",
        "authorized_key":[\
    \
        ],
        "host_key":[\
    \
        ]
      }
    }

### Description

Show data of last Linux installation

### Request limit

500 requests per 1 hour

### Output

linux (Object)

|                            |                                     |
| -------------------------- | ----------------------------------- |
| server_ip (String)         | Server main IP address              |
| server_ipv6_net (String)   | Server main IPv6 net address        |
| server_number (Integer)    | Server ID                           |
| dist (String)              | Distribution                        |
| @deprecated arch (Integer) | Architecture                        |
| lang (String)              | Language                            |
| active (Boolean)           | Linux installation status           |
| password (String)          | Linux installation password or null |
| authorized_key (Array)     | Authorized public SSH keys          |
| host_key (Array)           | Host keys                           |

### Errors

| Status | Code               | Description                                     |
| ------ | ------------------ | ----------------------------------------------- |
| 404    | SERVER_NOT_FOUND   | Server with id {server-number} not found        |
| 404    | BOOT_NOT_AVAILABLE | No boot configuration available for this server |

### Deprecations

|                                              |                                                                        |
| -------------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /boot/{server-ip}/linux/last | The main IPv4 address may be used alternatively to specify the server. |

## GET /boot/{server-number}/vnc

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/vnc


    {
      "vnc":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":[\
          "centOS-5.0",\
          "Fedora-6",\
          "openSUSE-10.2"\
        ],
        "@deprecated arch":[\
          64,\
          32\
        ],
        "lang":[\
          "de_DE",\
          "en_US"\
        ],
        "active":false,
        "password":null
      }
    }

### Description

Query boot options for the VNC installation

### Request limit

500 requests per 1 hour

### Output

vnc (Object)

|                                   |                                                             |
| --------------------------------- | ----------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                      |
| server_ipv6_net (String)          | Server main IPv6 net address                                |
| server_number (Integer)           | Server ID                                                   |
| dist (Array\|String)              | Array of available distributions or the active distributon  |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture |
| lang (Array\|String)              | Array of available languages or the active language         |
| active (Boolean)                  | Current VNC installation status                             |
| password (String)                 | Current VNC installation password or null                   |

### Errors

| Status | Code               | Description                                     |
| ------ | ------------------ | ----------------------------------------------- |
| 404    | SERVER_NOT_FOUND   | Server with id {server-number} not found        |
| 404    | BOOT_NOT_AVAILABLE | No boot configuration available for this server |

### Deprecations

|                                       |                                                                        |
| ------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /boot/{server-ip}/vnc | The main IPv4 address may be used alternatively to specify the server. |

## POST /boot/{server-number}/vnc

     curl -u "user:password" https://robot-ws.your-server.de/boot/321/vnc -d 'dist=centOS-5.0&lang=en_US'


    {
      "vnc":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":"centOS-5.0",
        "@deprecated arch":32,
        "lang":"en_US",
        "active":true,
        "password":"jEt0dtUvomlyOwRr"
      }
    }

### Description

Activate VNC installation

### Request limit

500 requests per 1 hour

### Input

| Name             | Description                          |
| ---------------- | ------------------------------------ |
| dist             | Distribution                         |
| @deprecated arch | Architecture (optional, default: 64) |
| lang             | Language                             |

### Output

vnc (Object)

|                            |                              |
| -------------------------- | ---------------------------- |
| server_ip (String)         | Server main IP address       |
| server_ipv6_net (String)   | Server main IPv6 net address |
| server_number (Integer)    | Server ID                    |
| dist (String)              | Distribution                 |
| @deprecated arch (Integer) | Architecture                 |
| lang (String)              | Language                     |
| active (Boolean)           | true                         |
| password (String)          | VNC installation password    |

### Errors

| Status | Code                   | Description                                                        |
| ------ | ---------------------- | ------------------------------------------------------------------ |
| 400    | INVALID_INPUT          | Invalid input parameters                                           |
| 404    | SERVER_NOT_FOUND       | Server with id {server-number} not found                           |
| 404    | BOOT_NOT_AVAILABLE     | No boot configuration available for this server                    |
| 500    | BOOT_ACTIVATION_FAILED | Activation of the VNC installation failed due to an internal error |

### Deprecations

|                                        |                                                                        |
| -------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated POST /boot/{server-ip}/vnc | The main IPv4 address may be used alternatively to specify the server. |

## DELETE /boot/{server-number}/vnc

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/vnc -X DELETE


    {
      "vnc":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":[\
          "centOS-5.0",\
          "Fedora-6",\
          "openSUSE-10.2"\
        ],
        "@deprecated arch":[\
          64,\
          32\
        ],
        "lang":[\
          "de_DE",\
          "en_US"\
        ],
        "active":false,
        "password":null
      }
    }

### Description

Deactivate VNC installation

### Request limit

500 requests per 1 hour

### Output

vnc (Object)

|                          |                                  |
| ------------------------ | -------------------------------- |
| server_ip (String)       | Server main IP address           |
| server_ipv6_net (String) | Server main IPv6 net address     |
| server_number (Integer)  | Server ID                        |
| dist (Array)             | Array of available distributions |
| @deprecated arch (Array) | Array of available architectures |
| lang (Array)             | Array of available languages     |
| active (Boolean)         | false                            |
| password (String)        | null                             |

### Errors

| Status | Code                     | Description                                                          |
| ------ | ------------------------ | -------------------------------------------------------------------- |
| 404    | SERVER_NOT_FOUND         | Server with id {server-number} not found                             |
| 404    | BOOT_NOT_AVAILABLE       | No boot configuration available for this server                      |
| 500    | BOOT_DEACTIVATION_FAILED | Deactivation of the VNC installation failed due to an internal error |

### Deprecations

|                                          |                                                                        |
| ---------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated DELETE /boot/{server-ip}/vnc | The main IPv4 address may be used alternatively to specify the server. |

## GET /boot/{server-number}/windows

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/windows


    {
      "windows":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":[\
          "standard"\
        ],
        "lang":[\
          "en",\
          "de"\
        ],
        "active":false,
        "password":null
      }
    }

### Description

Query boot options for the windows installation

### Request limit

500 requests per 1 hour

### Output

windows (Object)

|                                   |                                                             |
| --------------------------------- | ----------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                      |
| server_ipv6_net (String)          | Server main IPv6 net address                                |
| server_number (Integer)           | Server ID                                                   |
| dist (Array\|String)              | Array of available distributions or the active distributon  |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture |
| lang (Array\|String)              | Array of available languages or the active language         |
| active (Boolean)                  | Current Windows installation status                         |
| password (String)                 | Current Windows installation password or null               |

### Errors

| Status | Code                     | Description                                     |
| ------ | ------------------------ | ----------------------------------------------- |
| 404    | SERVER_NOT_FOUND         | Server with id {server-number} not found        |
| 404    | BOOT_NOT_AVAILABLE       | No boot configuration available for this server |
| 404    | WINDOWS_OUTDATED_VERSION | The windows version is not supported anymore    |

### Deprecations

|                                           |                                                                        |
| ----------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /boot/{server-ip}/windows | The main IPv4 address may be used alternatively to specify the server. |

## POST /boot/{server-number}/windows

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/windows -d 'lang=en'


    {
      "windows":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":"standard",
        "lang":"en",
        "active":true,
        "password":"jEt0dtUvomlyOwRr"
      }
    }

### Description

Activate Windows installation. You need to order the Windows addon for the server via the Robot webpanel first. After a reboot, the installation will start, and all data on the server will be deleted.

### Request limit

500 requests per 1 hour

### Input

| Name | Description |
| ---- | ----------- |
| lang | Language    |

### Output

windows (Object)

|                            |                               |
| -------------------------- | ----------------------------- |
| server_ip (String)         | Server main IP address        |
| server_ipv6_net (String)   | Server main IPv6 net address  |
| server_number (Integer)    | Server ID                     |
| dist (String)              | Distribution                  |
| @deprecated arch (Integer) | Architecture                  |
| lang (String)              | Language                      |
| active (Boolean)           | true                          |
| password (String)          | Windows installation password |

### Deprecations

|                                            |                                                                        |
| ------------------------------------------ | ---------------------------------------------------------------------- |
| @deprecated POST /boot/{server-ip}/windows | The main IPv4 address may be used alternatively to specify the server. |

### Errors

| Status | Code                     | Description                                                            |
| ------ | ------------------------ | ---------------------------------------------------------------------- |
| 400    | INVALID_INPUT            | Invalid input parameters                                               |
| 404    | SERVER_NOT_FOUND         | Server with id {server-number} not found                               |
| 404    | BOOT_NOT_AVAILABLE       | No boot configuration available for this server                        |
| 404    | WINDOWS_MISSING_ADDON    | No windows addon found                                                 |
| 404    | WINDOWS_OUTDATED_VERSION | The windows version is not supported anymore                           |
| 500    | BOOT_ACTIVATION_FAILED   | Activation of the windows installation failed due to an internal error |

## DELETE /boot/{server-number}/windows

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/windows -X DELETE


    {
      "windows":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":[\
          "standard"\
        ],
        "lang":[\
          "en",\
          "de"\
        ],
        "active":false,
        "password":null
      }
    }

### Description

Deactivate Windows installation

### Request limit

500 requests per 1 hour

### Output

windows (Object)

|                          |                                  |
| ------------------------ | -------------------------------- |
| server_ip (String)       | Server main IP address           |
| server_ipv6_net (String) | Server main IPv6 net address     |
| server_number (Integer)  | Server ID                        |
| dist (Array)             | Array of available distributions |
| @deprecated arch (Array) | Array of available architectures |
| lang (Array)             | Array of available languages     |
| active (Boolean)         | false                            |
| password (String)        | null                             |

### Errors

| Status | Code                     | Description                                                              |
| ------ | ------------------------ | ------------------------------------------------------------------------ |
| 404    | SERVER_NOT_FOUND         | Server with id {server-number} not found                                 |
| 404    | BOOT_NOT_AVAILABLE       | No boot configuration available for this server                          |
| 404    | WINDOWS_OUTDATED_VERSION | The windows version is not supported anymore                             |
| 500    | BOOT_DEACTIVATION_FAILED | Deactivation of the windows installation failed due to an internal error |

### Deprecations

|                                              |                                                                        |
| -------------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated DELETE /boot/{server-ip}/windows | The main IPv4 address may be used alternatively to specify the server. |

## GET /boot/{server-number}/plesk

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/plesk


    {
      "plesk":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":[\
          "CentOS 5.4 minimal",\
          "Debian 7.8 minimal"\
        ],
        "@deprecated arch":[\
          64,\
          32\
        ],
        "lang":[\
          "en",\
          "de"\
        ],
        "active":false,
        "password":null,
        "hostname":null
      }
    }

### Description

Query boot options for the Plesk installation

### Request limit

500 requests per 1 hour

### Output

plesk (Object)

|                                   |                                                             |
| --------------------------------- | ----------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                      |
| server_ipv6_net (String)          | Server main IPv6 net address                                |
| server_number (Integer)           | Server ID                                                   |
| dist (Array\|String)              | Array of available distributions or the active distributon  |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture |
| lang (Array\|String)              | Array of available languages or the active language         |
| active (Boolean)                  | Current Plesk installation status                           |
| password (String)                 | Current Plesk installation password or null                 |
| hostname (String)                 | Current Plesk installation hostname or null                 |

### Errors

| Status | Code                     | Description                                     |
| ------ | ------------------------ | ----------------------------------------------- |
| 404    | SERVER_NOT_FOUND         | Server with id {server-number} not found        |
| 404    | BOOT_NOT_AVAILABLE       | No boot configuration available for this server |
| 404    | WINDOWS_OUTDATED_VERSION | The windows version is not supported anymore    |

### Deprecations

|                                         |                                                                        |
| --------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /boot/{server-ip}/plesk | The main IPv4 address may be used alternatively to specify the server. |

## POST /boot/{server-number}/plesk

    curl -u "user:password" https://robot-ws.your-server.de/boot.yaml/321/plesk -d 'dist=CentOS 5.4 minimal&lang=de&hostname=plesk.testen.de'


    {
      "plesk":{
        "server_ip":"213.239.217.200",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":"CentOS 5.4 minimal",
        "@deprecated arch":32,
        "lang":"de",
        "active":true,
        "password":"jEt0dtUvomlyOwRr",
        "hostname":"plesk.testen.de"
      }
    }

### Description

Activate Plesk installation

### Request limit

500 requests per 1 hour

### Input

| Name             | Description                          |
| ---------------- | ------------------------------------ |
| dist             | Distribution                         |
| @deprecated arch | Architecture (optional, default: 64) |
| lang             | Language                             |
| hostname         | Hostname                             |

### Output

plesk (Object)

|                            |                              |
| -------------------------- | ---------------------------- |
| server_ip (String)         | Server main IP address       |
| server_ipv6_net (String)   | Server main IPv6 net address |
| server_number (Integer)    | Server ID                    |
| dist (String)              | Distribution                 |
| @deprecated arch (Integer) | Architecture                 |
| lang (String)              | Language                     |
| active (Boolean)           | true                         |
| password (String)          | Plesk installation password  |
| hostname (String)          | Plesk installation hostname  |

### Errors

| Status | Code                     | Description                                                          |
| ------ | ------------------------ | -------------------------------------------------------------------- |
| 400    | INVALID_INPUT            | Invalid input parameters                                             |
| 404    | SERVER_NOT_FOUND         | Server with id {server-number} not found                             |
| 404    | BOOT_NOT_AVAILABLE       | No boot configuration available for this server                      |
| 404    | PLESK_MISSING_ADDON      | No plesk addon found                                                 |
| 404    | WINDOWS_OUTDATED_VERSION | The windows version is not supported anymore                         |
| 500    | BOOT_ACTIVATION_FAILED   | Activation of the Plesk installation failed due to an internal error |

### Deprecations

|                                          |                                                                        |
| ---------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated POST /boot/{server-ip}/plesk | The main IPv4 address may be used alternatively to specify the server. |

## DELETE /boot/{server-number}/plesk

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/plesk -X DELETE


    {
      "plesk":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":[\
          "CentOS 5.4 minimal",\
          "Debian 7.8 minimal"\
        ],
        "@deprecated arch":[\
          64,\
          32\
        ],
        "lang":[\
          "en",\
          "de"\
        ],
        "active":false,
        "password":null,
        "hostname":null
      }
    }

### Description

Deactivate Plesk installation

### Request limit

500 requests per 1 hour

### Output

plesk (Object)

|                          |                                  |
| ------------------------ | -------------------------------- |
| server_ip (String)       | Server main IP address           |
| server_ipv6_net (String) | Server main IPv6 net address     |
| server_number (Integer)  | Server ID                        |
| dist (Array)             | Array of available distributions |
| @deprecated arch (Array) | Array of available architectures |
| lang (Array)             | Array of available languages     |
| active (Boolean)         | false                            |
| password (String)        | null                             |
| hostname (String)        | null                             |

### Errors

| Status | Code                     | Description                                                            |
| ------ | ------------------------ | ---------------------------------------------------------------------- |
| 404    | SERVER_NOT_FOUND         | Server with id {server-number} not found                               |
| 404    | BOOT_NOT_AVAILABLE       | No boot configuration available for this server                        |
| 404    | WINDOWS_OUTDATED_VERSION | The windows version is not supported anymore                           |
| 500    | BOOT_DEACTIVATION_FAILED | Deactivation of the Plesk installation failed due to an internal error |

### Deprecations

|                                            |                                                                        |
| ------------------------------------------ | ---------------------------------------------------------------------- |
| @deprecated DELETE /boot/{server-ip}/plesk | The main IPv4 address may be used alternatively to specify the server. |

## GET /boot/{server-number}/cpanel

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/cpanel


    {
      "cpanel":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":[\
          "CentOS 5.6 + cPanel"\
        ],
        "@deprecated arch":[\
          64\
        ],
        "lang":[\
          "en"\
        ],
        "active":false,
        "password":null,
        "hostname":null
      }
    }

### Description

Query boot options for the cPanel installation

### Request limit

500 requests per 1 hour

### Output

cpanel (Object)

|                                   |                                                             |
| --------------------------------- | ----------------------------------------------------------- |
| server_ip (String)                | Server main IP address                                      |
| server_ipv6_net (String)          | Server main IPv6 net address                                |
| server_number (Integer)           | Server ID                                                   |
| dist (Array\|String)              | Array of available distributions or the active distributon  |
| @deprecated arch (Array\|Integer) | Array of available architectures or the active architecture |
| lang (Array\|String)              | Array of available languages or the active language         |
| active (Boolean)                  | Current cPanel installation status                          |
| password (String)                 | Current cPanel installation password or null                |
| hostname (String)                 | Current cPanel installation hostname or null                |

### Errors

| Status | Code               | Description                                     |
| ------ | ------------------ | ----------------------------------------------- |
| 404    | SERVER_NOT_FOUND   | Server with id {server-number} not found        |
| 404    | BOOT_NOT_AVAILABLE | No boot configuration available for this server |

### Deprecations

|                                          |                                                                        |
| ---------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /boot/{server-ip}/cpanel | The main IPv4 address may be used alternatively to specify the server. |

## POST /boot/{server-number}/cpanel

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/cpanel -d 'dist=CentOS 5.6 + cPanel&lang=en&hostname=cpanel.testen.de'


    {
      "cpanel":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":"CentOS 5.6 + cPanel",
        "@deprecated arch":64,
        "lang":"en",
        "active":true,
        "password":"ie8Nhz6R",
        "hostname":"cpanel.testen.de"
      }
    }

### Description

Activate cPanel installation

### Request limit

500 requests per 1 hour

### Input

| Name             | Description                          |
| ---------------- | ------------------------------------ |
| dist             | Distribution                         |
| @deprecated arch | Architecture (optional, default: 64) |
| lang             | Language                             |
| hostname         | Hostname                             |

### Output

cpanel (Object)

|                            |                              |
| -------------------------- | ---------------------------- |
| server_ip (String)         | Server main IP address       |
| server_ipv6_net (String)   | Server main IPv6 net address |
| server_number (Integer)    | Server ID                    |
| dist (String)              | Distribution                 |
| @deprecated arch (Integer) | Architecture                 |
| lang (String)              | Language                     |
| active (Boolean)           | true                         |
| password (String)          | cPanel installation password |
| hostname (String)          | cPanel installation hostname |

### Errors

| Status | Code                   | Description                                                           |
| ------ | ---------------------- | --------------------------------------------------------------------- |
| 400    | INVALID_INPUT          | Invalid input parameters                                              |
| 404    | SERVER_NOT_FOUND       | Server with id {server-number} not found                              |
| 404    | BOOT_NOT_AVAILABLE     | No boot configuration available for this server                       |
| 404    | CPANEL_MISSING_ADDON   | No cPanel addon found                                                 |
| 500    | BOOT_ACTIVATION_FAILED | Activation of the cPanel installation failed due to an internal error |

### Deprecations

|                                           |                                                                        |
| ----------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated POST /boot/{server-ip}/cpanel | The main IPv4 address may be used alternatively to specify the server. |

## DELETE /boot/{server-number}/cpanel

    curl -u "user:password" https://robot-ws.your-server.de/boot/321/cpanel -X DELETE


    {
      "cpanel":{
        "server_ip":"123.123.123.123",
        "server_ipv6_net":"2a01:4f8:111:4221::",
        "server_number":321,
        "dist":[\
          "CentOS 5.6 + cPanel"\
        ],
        "@deprecated arch":[\
          64\
        ],
        "lang":[\
          "en"\
        ],
        "active":false,
        "password":null,
        "hostname":null
      }
    }

### Description

Deactivate cPanel installation

### Request limit

500 requests per 1 hour

### Output

plesk (Object)

|                          |                                  |
| ------------------------ | -------------------------------- |
| server_ip (String)       | Server main IP address           |
| server_ipv6_net (String) | Server main IPv6 net address     |
| server_number (Integer)  | Server ID                        |
| dist (Array)             | Array of available distributions |
| @deprecated arch (Array) | Array of available architectures |
| lang (Array)             | Array of available languages     |
| active (Boolean)         | false                            |
| password (String)        | null                             |
| hostname (String)        | null                             |

### Errors

| Status | Code                     | Description                                                             |
| ------ | ------------------------ | ----------------------------------------------------------------------- |
| 404    | SERVER_NOT_FOUND         | Server with id {server-number} not found                                |
| 404    | BOOT_NOT_AVAILABLE       | No boot configuration available for this server                         |
| 500    | BOOT_DEACTIVATION_FAILED | Deactivation of the cPanel installation failed due to an internal error |

### Deprecations

|                                             |                                                                        |
| ------------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated DELETE /boot/{server-ip}/cpanel | The main IPv4 address may be used alternatively to specify the server. |

# Reverse DNS

## GET /rdns

    curl -u "user:password" https://robot-ws.your-server.de/rdns


    [\
      {\
        "rdns":{\
          "ip":"123.123.123.123",\
          "ptr":"testen.de"\
        }\
      },\
      {\
        "rdns":{\
          "ip":"124.124.124.124",\
          "ptr":"your-server.de"\
        }\
      }\
    ]

### Description

Query all rDNS entries

### Request limit

500 requests per 1 hour

### Input (optional)

| Name      | Description                                                                   |
| --------- | ----------------------------------------------------------------------------- |
| server_ip | Server main IP address; show only reverse DNS entries assigned to this server |

### Output

(Array)rdns (Object)

|              |            |
| ------------ | ---------- |
| ip (String)  | IP address |
| ptr (String) | PTR record |

### Errors

| Status | Code      | Description                  |
| ------ | --------- | ---------------------------- |
| 404    | NOT_FOUND | No reverse DNS entries found |

## GET /rdns/{ip}

    curl -u "user:password" https://robot-ws.your-server.de/rdns/123.123.123.123


    {
      "rdns":{
        "ip":"123.123.123.123",
        "ptr":"testen.de"
      }
    }

### Description

Query the current reverse DNS entry for one IP address

### Request limit

500 requests per 1 hour

### Output

rdns (Object)

|              |            |
| ------------ | ---------- |
| ip (String)  | IP address |
| ptr (String) | PTR record |

### Errors

| Status | Code           | Description                                      |
| ------ | -------------- | ------------------------------------------------ |
| 404    | IP_NOT_FOUND   | The IP address {ip} was not found                |
| 404    | RDNS_NOT_FOUND | The IP address {ip} has no reverse DNS entry yet |

## PUT /rdns/{ip}

    curl -u "user:password" https://robot-ws.your-server.de/rdns/123.123.123.123 -d ptr=testen.de -X PUT


    {
      "rdns":{
        "ip":"123.123.123.123",
        "ptr":"testen.de"
      }
    }

### Description

Create new reverse DNS entry for one IP address. Once the reverse DNS entry is successfully created, the status code 201 CREATED is returned.

### Request limit

500 requests per 1 hour

### Input

| Name | Description |
| ---- | ----------- |
| ptr  | PTR record  |

### Output

rdns (Object)

|              |            |
| ------------ | ---------- |
| ip (String)  | IP address |
| ptr (String) | PTR record |

### Errors

| Status | Code                | Description                                                    |
| ------ | ------------------- | -------------------------------------------------------------- |
| 400    | INVALID_INPUT       | Invalid input parameters                                       |
| 404    | IP_NOT_FOUND        | The IP address {ip} was not found                              |
| 409    | RDNS_ALREADY_EXISTS | There is already an existing reverse DNS entry                 |
| 500    | RDNS_CREATE_FAILED  | Creating the reverse DNS entry failed due to an internal error |

## POST /rdns/{ip}

    curl -u "user:password" https://robot-ws.your-server.de/rdns/123.123.123.123 -d ptr=testen.de


    {
      "rdns":{
        "ip":"123.123.123.123",
        "ptr":"testen.de"
      }
    }

### Description

Update/create a reverse DNS entry for one IP. Once the reverse DNS entry is successfully created, the status code is set to 201 created. On succesfull updates, the status code is 200 OK.

### Request limit

500 requests per 1 hour

### Input

| Name | Description |
| ---- | ----------- |
| ptr  | PTR record  |

### Output

rdns (Object)

|              |            |
| ------------ | ---------- |
| ip (String)  | IP address |
| ptr (String) | PTR record |

### Errors

| Status | Code               | Description                                                    |
| ------ | ------------------ | -------------------------------------------------------------- |
| 400    | INVALID_INPUT      | Invalid input parameters                                       |
| 404    | IP_NOT_FOUND       | The IP address {ip} was not found                              |
| 500    | RDNS_CREATE_FAILED | Creating the reverse DNS entry failed due to an internal error |
| 500    | RDNS_UPDATE_FAILED | Updating the reverse DNS entry failed due to an internal error |

## DELETE /rdns/{ip}

    curl -u "user:password" https://robot-ws.your-server.de/rdns/123.123.123.123 -X DELETE

### Description

Delete reverse DNS entry for one IP

### Request limit

500 requests per 1 hour

### Output

No output

### Errors

| Status | Code               | Description                                                    |
| ------ | ------------------ | -------------------------------------------------------------- |
| 404    | IP_NOT_FOUND       | The IP address {ip} was not found                              |
| 500    | RDNS_DELETE_FAILED | Deleting the reverse DNS entry failed due to an internal error |
| 500    | RDNS_UPDATE_FAILED | Updating the reverse DNS entry failed due to an internal error |

# Traffic

## POST /traffic

> Query traffic data for one IP

    curl -u "user:password" https://robot-ws.your-server.de/traffic \
      --data-urlencode 'type=month' \
      --data-urlencode 'from=2010-09-01' \
      --data-urlencode 'to=2010-09-31' \
      --data-urlencode 'ip=123.123.123.123'


    {
      "traffic":{
        "type":"month",
        "from":"2010-09-01",
        "to":"2010-09-31",
        "data":{
          "123.123.123.123":{
            "in":0.2874,
            "out":0.0481,
            "sum":0.3355
          }
        }
      }
    }

> Query traffic data for multiple IPs

    curl -u "user:password" https://robot-ws.your-server.de/traffic \
      --data-urlencode 'type=month' \
      --data-urlencode 'from=2010-09-01' \
      --data-urlencode 'to=2010-09-31' \
      --data-urlencode 'ip[]=123.123.123.123' \
      --data-urlencode 'ip[]=124.124.124.124'


    {
      "traffic":{
        "type":"month",
        "from":"2010-09-01",
        "to":"2010-09-31",
        "data":{
          "123.123.123.123":{
            "in":0.2874,
            "out":0.0481,
            "sum":0.3355
          },
          "124.124.124.124":{
            "in":0.2874,
            "out":0.0481,
            "sum":0.3355
          }
        }
      }
    }

> Query traffic data for subnet

    curl -u "user:password" https://robot-ws.your-server.de/traffic \
      --data-urlencode 'type=month' \
      --data-urlencode 'from=2010-09-01' \
      --data-urlencode 'to=2010-09-31' \
      --data-urlencode 'subnet=2a01:4f8:61:41a2::'


    {
      "traffic":{
        "type":"month",
        "from":"2010-09-01",
        "to":"2010-09-31",
        "data":{
          "2a01:4f8:61:41a2::\/64":{
            "in":0.2874,
            "out":0.0481,
            "sum":0.3355
          }
        }
      }
    }

> Query traffic data grouped by days for one IP

    curl -u "user:password" https://robot-ws.your-server.de/traffic \
      --data-urlencode 'type=month' \
      --data-urlencode 'from=2019-01-01' \
      --data-urlencode 'to=2019-01-07' \
      --data-urlencode 'ip=123.123.123.123' \
      --data-urlencode 'single_values=true'


    {
      "traffic":{
        "type":"month",
        "from":"2019-01-01",
        "to":"2019-01-07",
        "data":{
          "123.123.123.123":{
            "01":{
              "in":0.0023,
              "out":0.0102,
              "sum":0.0125
            },
            "02":{
              "in":229.7502,
              "out":10.7187,
              "sum":240.4689
            },
            "03":{
              "in":97.8517,
              "out":1.53,
              "sum":99.3817
            },
            "04":{
              "in":191.0187,
              "out":0.153,
              "sum":191.1717
            },
            "05":{
              "in":0.0021,
              "out":0.0022,
              "sum":0.0043
            },
            "06":{
              "in":0,
              "out":0.0021,
              "sum":0.0021
            },
            "07":{
              "in":0,
              "out":0,
              "sum":0
            }
          }
        }
      }
    }

### Description

Query traffic data of IPs and subnets. There are three query types: "day", "month" and "year". With "day" you can query hourly aggregated traffic data within a day. With "month" you can query daily aggregated data within a month. And with "year" it is possible to get monthly aggregated data within a year.

Please note that the traffic data is only available once the specified hour, day or month has already passed.

The interval is given with the parameters "from" and "to" with the following syntax:

- Query type "day": \[YYYY\]-\[MM\]-\[DD\]T\[H\], e.g. 2010-09-01T00
- Query type "month": \[YYYY\]-\[MM\]-\[DD\], e.g. 2010-09-01
- Query type "year": \[YYYY\]-\[MM\], e.g. 2010-09

When using the query type "day", you must specify a date combined with a time value (hour). The hour value is separated from the date with the letter "T". When using the query type "month", you must specify a date without the time value. With query type "year", you must additionally leave out the day value.

IP addresses or subnets without traffic data are omitted in the response.

Using the parameter "single_values" it is possible to get the traffic data grouped by hours, days or month over the specified interval. For type "day" the data is grouped by hours, for type "month" by days and for type "year" by months.

### Request limit

200 requests per 1 hour

### Input

| Name          | Description                                                                                                                     |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| ip\[\]        | One or more IP addresses                                                                                                        |
| subnet\[\]    | One or more subnet addresses                                                                                                    |
| from          | Date/Time from                                                                                                                  |
| to            | Date/Time to                                                                                                                    |
| type          | Type of traffic query                                                                                                           |
| single_values | If set to "true" the traffic data is returned not as a sum over the whole interval but grouped by hour, day or month (optional) |

### Output

Output without parameter "single_values"

traffic (Object)type (String)Traffic query typefrom (String)Date/Time fromto (String)Date/Time todata (Object)<IP address> (Object)

|              |                  |
| ------------ | ---------------- |
| in (Number)  | traffic in (GB)  |
| out (Number) | traffic out (GB) |
| sum (Number) | traffic sum (GB) |

Output with parameter "single_values"

traffic (Object)type (String)Traffic query typefrom (String)Date/Time fromto (String)Date/Time todata (Object)<IP address> (Object)<interval> (Object)

|              |                  |
| ------------ | ---------------- |
| in (Number)  | traffic in (GB)  |
| out (Number) | traffic out (GB) |
| sum (Number) | traffic sum (GB) |

### Errors

| Status | Code           | Description                                   |
| ------ | -------------- | --------------------------------------------- |
| 400    | INVALID_INPUT  | Invalid input parameters                      |
| 404    | NOT_FOUND      | No IP addresses or subnets found              |
| 500    | INTERNAL_ERROR | Traffic query failed due to an internal error |

# SSH keys

## GET /key

    curl -u "user:password" https://robot-ws.your-server.de/key


    [\
      {\
        "key":{\
          "name":"key1",\
          "fingerprint":"56:29:99:a4:5d:ed:ac:95:c1:f5:88:82:90:5d:dd:10",\
          "type":"ECDSA",\
          "size":521,\
          "data":"ecdsa-sha2-nistp521 AAAAE2VjZHNh ...",\
          "created_at":"2021-12-31 23:59:59"\
        }\
      },\
      {\
        "key":{\
          "name":"key2",\
          "fingerprint":"15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb",\
          "type":"ED25519",\
          "size":256,\
          "data":"ssh-ed25519 AAAAC3NzaC1 ...",\
          "created_at":"2021-12-31 23:59:59"\
        }\
      }\
    ]

### Description

Query all SSH keys

### Request limit

500 requests per 1 hour

### Output

(Array)key (Object)

|                      |                            |
| -------------------- | -------------------------- |
| name (String)        | Key name                   |
| fingerprint (String) | Key fingerprint            |
| type (String)        | Key algorithm type         |
| size (Integer)       | Key size in bits           |
| data (String)        | Key data in OpenSSH format |
| created_at (Date)    | Key creation date          |

### Errors

| Status | Code      | Description   |
| ------ | --------- | ------------- |
| 404    | NOT_FOUND | No keys found |

## POST /key

    curl -u "user:password" https://robot-ws.your-server.de/key \
    --data-urlencode 'name=NewKey' \
    --data-urlencode 'data=ssh-rsa AAAAB3NzaC1yc+...'


    {
      "key":{
        "name":"NewKey",
        "fingerprint":"cb:8b:ef:a7:fe:04:87:3f:e5:55:cd:12:e3:e8:9f:99",
        "type":"RSA",
        "size":8192,
        "data":"ssh-rsa AAAAB3NzaC1yc+...",
        "created_at":"2021-12-31 23:59:59"
      }
    }

### Description

Add a new SSH key. Once the key is successfully added, the status code 201 CREATED is returned.

### Request limit

200 requests per 1 hour

### Input

| Name | Description                            |
| ---- | -------------------------------------- |
| name | SSH key name                           |
| data | SSH key data in OpenSSH or SSH2 format |

### Output

key (Object)

|                      |                            |
| -------------------- | -------------------------- |
| name (String)        | Key name                   |
| fingerprint (String) | Key fingerprint            |
| type (String)        | Key algorithm type         |
| size (Integer)       | Key size in bits           |
| data (String)        | Key data in OpenSSH format |
| created_at (Date)    | Key creation date          |

### Errors

| Status | Code               | Description                                    |
| ------ | ------------------ | ---------------------------------------------- |
| 400    | INVALID_INPUT      | Invalid input parameters                       |
| 409    | KEY_ALREADY_EXISTS | The supplied key already exists                |
| 500    | KEY_CREATE_FAILED  | Adding the key failed due to an internal error |

## GET /key/{fingerprint}

    curl -u "user:password" https://robot-ws.your-server.de/key/15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb


    {
      "key":{
        "name":"key2",
        "fingerprint":"15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb",
        "type":"ED25519",
        "size":256,
        "data":"ssh-ed25519 AAAAC3NzaC1 ...",
        "created_at":"2021-12-31 23:59:59"
      }
    }

### Description

Query a specific SSH key

### Request limit

500 requests per 1 hour

### Output

key (Object)

|                      |                            |
| -------------------- | -------------------------- |
| name (String)        | Key name                   |
| fingerprint (String) | Key fingerprint            |
| type (String)        | Key algorithm type         |
| size (Integer)       | Key size in bits           |
| data (String)        | Key data in OpenSSH format |
| created_at (Date)    | Key creation date          |

### Errors

| Status | Code      | Description   |
| ------ | --------- | ------------- |
| 404    | NOT_FOUND | Key not found |

## POST /key/{fingerprint}

    curl -u "user:password" https://robot-ws.your-server.de/key/15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb \
    --data-urlencode 'name=MyTestKey'


    {
      "key":{
        "name":"MyTestKey",
        "fingerprint":"15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb",
        "type":"ED25519",
        "size":256,
        "data":"ssh-ed25519 AAAAC3NzaC1 ...",
        "created_at":"2021-12-31 23:59:59"
      }
    }

### Description

Update the key name

### Request limit

200 requests per 1 hour

### Input

| Name | Description  |
| ---- | ------------ |
| name | SSH key name |

### Output

key (Object)

|                      |                            |
| -------------------- | -------------------------- |
| name (String)        | Key name                   |
| fingerprint (String) | Key fingerprint            |
| type (String)        | Key algorithm type         |
| size (Integer)       | Key size in bits           |
| data (String)        | Key data in OpenSSH format |
| created_at (Date)    | Key creation date          |

### Errors

| Status | Code              | Description                                           |
| ------ | ----------------- | ----------------------------------------------------- |
| 400    | INVALID_INPUT     | Invalid input parameters                              |
| 404    | NOT_FOUND         | Key not found                                         |
| 500    | KEY_UPDATE_FAILED | Updating the key name failed due do an internal error |

## DELETE /key/{fingerprint}

    curl -u "user:password" https://robot-ws.your-server.de/key/15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb -X DELETE

### Description

Remove public key

### Request limit

200 requests per 1 hour

### Output

No output

### Errors

| Status | Code              | Description                                      |
| ------ | ----------------- | ------------------------------------------------ |
| 404    | NOT_FOUND         | Key not found                                    |
| 500    | KEY_DELETE_FAILED | Deleting the key failed due do an internal error |

# Server ordering

## Activation

To use the Robot webservice for ordering servers, please activate this function in your Robot administrative interface first via "Administration; Settings; Web Service Settings; Ordering".

## Notes

As of 28 March 2022, the listed prices no longer include the "Primary IPv4" addon. By default, we deploy servers as IPv6-only servers without an IPv4 address. If you actively select the addon "primary_ipv4" when you are making your order, we will deliver the server to you with an IPv4 address. Addon prices are listed separately.

## GET /order/server/product

    curl -u "user:password" https://robot-ws.your-server.de/order/server/product


    [\
      {\
        "product":{\
          "id":"EX60",\
          "name":"Dedicated Server EX60",\
          "description":[\
            "Intel\u00ae Core\u2122 i7-920 Quad-Core",\
            "48 GB DDR3 RAM",\
            "2 x 2 TB SATA 3 Gb\/s Enterprise HDD; 7200 rpm(Software-RAID 1)",\
            "1 Gbit\/s bandwidth"\
          ],\
          "traffic":"30 TB",\
          "dist":[\
            "Rescue system",\
            "CentOS 6.6 minimal",\
            "CentOS 7.0 minimal",\
            "Debian 7.7 LAMP",\
            "Debian 7.7 minimal",\
            "openSUSE 13.2 minimal",\
            "Ubuntu 14.04.1 LTS minimal",\
            "Ubuntu 14.10 minimal"\
          ],\
          "@deprecated arch":[\
            64,\
            32\
          ],\
          "lang":[\
            "en"\
          ],\
          "location":[\
            "FSN1",\
            "NBG1"\
          ],\
          "prices":[\
            {\
              "location":"FSN1",\
              "price":{\
                "net":"49.58",\
                "gross":"49.58",\
                "hourly_net":"0.0795",\
                "hourly_gross":"0.0795"\
              },\
              "price_setup":{\
                "net":"0.00",\
                "gross":"0.00"\
              }\
            },\
            {\
              "location":"NBG1",\
              "price":{\
                "net":"49.58",\
                "gross":"49.58",\
                "hourly_net":"0.0795",\
                "hourly_gross":"0.0795"\
              },\
              "price_setup":{\
                "net":"0.00",\
                "gross":"0.00"\
              }\
            }\
          ],\
          "orderable_addons":[\
            {\
              "id":"primary_ipv4",\
              "name":"Primary IPv4",\
              "min":0,\
              "max":1,\
              "prices":[\
                {\
                  "location":"FSN1",\
                  "price":{\
                    "net":"1.7000",\
                    "gross":"1.7000",\
                    "hourly_net":"0.0027",\
                    "hourly_gross":"0.0027"\
                  },\
                  "price_setup":{\
                    "net":"0.0000",\
                    "gross":"0.0000"\
                  }\
                },\
                {\
                  "location":"NBG1",\
                  "price":{\
                    "net":"1.7000",\
                    "gross":"1.7000",\
                    "hourly_net":"0.0027",\
                    "hourly_gross":"0.0027"\
                  },\
                  "price_setup":{\
                    "net":"0.0000",\
                    "gross":"0.0000"\
                  }\
                }\
              ]\
            }\
          ]\
        }\
      },\
      {\
        "product":{\
          "id":"EX40",\
          "name":"Dedicated Server EX40",\
          "description":[\
            "Intel\u00ae Core\u2122 i7-4770 Quad-Core Haswell",\
            "32 GB DDR3 RAM",\
            "2 x 2 TB SATA 6 Gb\/s Enterprise HDD; 7200 rpm(Software-RAID 1)",\
            "1 Gbit\/s bandwidth"\
          ],\
          "traffic":"30 TB",\
          "dist":[\
            "Rescue system",\
            "CentOS 6.6 minimal",\
            "CentOS 7.0 minimal",\
            "Debian 7.7 LAMP",\
            "Debian 7.7 minimal",\
            "openSUSE 13.2 minimal",\
            "Ubuntu 14.04.1 LTS minimal",\
            "Ubuntu 14.10 minimal"\
          ],\
          "@deprecated arch":[\
            64,\
            32\
          ],\
          "lang":[\
            "en"\
          ],\
          "location":[\
            "FSN1",\
            "NBG1"\
          ],\
          "prices":[\
            {\
              "location":"FSN1",\
              "price":{\
                "net":"84.03",\
                "gross":"84.03",\
                "hourly_net":"0.1347",\
                "hourly_gross":"0.1347"\
              },\
              "price_setup":{\
                "net":"41.18",\
                "gross":"41.18"\
              }\
            },\
            {\
              "location":"NBG1",\
              "price":{\
                "net":"84.03",\
                "gross":"84.03",\
                "hourly_net":"0.1347",\
                "hourly_gross":"0.1347"\
              },\
              "price_setup":{\
                "net":"41.18",\
                "gross":"41.18"\
              }\
            }\
          ],\
          "orderable_addons":[\
            {\
              "id":"primary_ipv4",\
              "name":"Primary IPv4",\
              "min":0,\
              "max":1,\
              "prices":[\
                {\
                  "location":"FSN1",\
                  "price":{\
                    "net":"1.7000",\
                    "gross":"1.7000",\
                    "hourly_net":"0.0027",\
                    "hourly_gross":"0.0027"\
                  },\
                  "price_setup":{\
                    "net":"0.0000",\
                    "gross":"0.0000"\
                  }\
                },\
                {\
                  "location":"NBG1",\
                  "price":{\
                    "net":"1.7000",\
                    "gross":"1.7000",\
                    "hourly_net":"0.0027",\
                    "hourly_gross":"0.0027"\
                  },\
                  "price_setup":{\
                    "net":"0.0000",\
                    "gross":"0.0000"\
                  }\
                }\
              ]\
            }\
          ]\
        }\
      }\
    ]

### Description

Product overview of currently offered standard server products

### Request limit

500 requests per 1 hour

### Input (optional)

| Name            | Description           |
| --------------- | --------------------- |
| min_price       | Minimum monthly price |
| max_price       | Maximum monthly price |
| min_price_setup | Minimum one time fee  |
| max_price_setup | Maximum one time fee  |
| location        | The desired location  |

### Output

(Array)product (Object)id (String)Product IDname (String)Product namedescription (Array)Textual descriptiontraffic (String)Free traffic quotadist (Array)Available distributions@deprecated arch (Array)Available distribution architectureslang (Array)Available distribution languageslocation (Array)Available locationsprices (Array)(Object)location (String)Locationprice (Object)

|                       |                                                                                 |
| --------------------- | ------------------------------------------------------------------------------- |
| net (String)          | Monthly price in euros                                                          |
| gross (String)        | Monthly price in euros with VAT                                                 |
| hourly_net (String)   | Hourly price in euros, if the product is billed hourly, null otherwise          |
| hourly_gross (String) | Hourly price in euros with VAT, if the product is billed hourly, null otherwise |

price_setup (Object)

|                |                                |
| -------------- | ------------------------------ |
| net (String)   | One time fee in euros          |
| gross (String) | One time fee in euros with VAT |

orderable_addons (Array)(Object)id (String)Addon IDname (String)Addon namelocation (String)Locationmin (Integer)Minimum orderable amountmax (Integer)Maximum orderable amountprices (Array)(Object)location (String)Locationprice (Object)

|                       |                                                                                 |
| --------------------- | ------------------------------------------------------------------------------- |
| net (String)          | Monthly price in euros                                                          |
| gross (String)        | Monthly price in euros with VAT                                                 |
| hourly_net (String)   | Hourly price in euros, if the product is billed hourly, null otherwise          |
| hourly_gross (String) | Hourly price in euros with VAT, if the product is billed hourly, null otherwise |

price_setup (Object)

|                |                                |
| -------------- | ------------------------------ |
| net (String)   | One time fee in euros          |
| gross (String) | One time fee in euros with VAT |

### Errors

| Status | Code      | Description       |
| ------ | --------- | ----------------- |
| 404    | NOT_FOUND | No products found |

## GET /order/server/product/{product-id}

    curl -u "user:password" https://robot-ws.your-server.de/order/server/product/EX40


    {
      "product":{
        "id":"EX40",
        "name":"Dedicated Server EX40",
        "description":[\
          "Intel\u00ae Core\u2122 i7-4770 Quad-Core Haswell",\
          "32 GB DDR3 RAM",\
          "2 x 2 TB SATA 6 Gb\/s Enterprise HDD; 7200 rpm(Software-RAID 1)",\
          "1 Gbit\/s bandwidth"\
        ],
        "traffic":"30 TB",
        "dist":[\
          "Rescue system",\
          "CentOS 6.6 minimal",\
          "CentOS 7.0 minimal",\
          "Debian 7.7 LAMP",\
          "Debian 7.7 minimal",\
          "openSUSE 13.2 minimal",\
          "Ubuntu 14.04.1 LTS minimal",\
          "Ubuntu 14.10 minimal"\
        ],
        "@deprecated arch":[\
          64,\
          32\
        ],
        "lang":[\
          "en"\
        ],
        "location":[\
          "FSN1",\
          "NBG1"\
        ],
        "prices":[\
          {\
            "location":"FSN1",\
            "price":{\
              "net":"84.03",\
              "gross":"84.03",\
              "hourly_net":"0.1347",\
              "hourly_gross":"0.1347"\
            },\
            "price_setup":{\
              "net":"41.18",\
              "gross":"41.18"\
            }\
          },\
          {\
            "location":"NBG1",\
            "price":{\
              "net":"84.03",\
              "gross":"84.03",\
              "hourly_net":"0.1347",\
              "hourly_gross":"0.1347"\
            },\
            "price_setup":{\
              "net":"41.18",\
              "gross":"41.18"\
            }\
          }\
        ],
        "orderable_addons":[\
          {\
            "id":"primary_ipv4",\
            "name":"Primary IPv4",\
            "min":0,\
            "max":1,\
            "prices":[\
              {\
                "location":"FSN1",\
                "price":{\
                  "net":"1.7000",\
                  "gross":"1.7000",\
                  "hourly_net":"0.0027",\
                  "hourly_gross":"0.0027"\
                },\
                "price_setup":{\
                  "net":"0.0000",\
                  "gross":"0.0000"\
                }\
              },\
              {\
                "location":"NBG1",\
                "price":{\
                  "net":"1.7000",\
                  "gross":"1.7000",\
                  "hourly_net":"0.0027",\
                  "hourly_gross":"0.0027"\
                },\
                "price_setup":{\
                  "net":"0.0000",\
                  "gross":"0.0000"\
                }\
              }\
            ]\
          }\
        ]
      }
    }

### Description

Query a specific server product

### Request limit

500 requests per 1 hour

### Output

product (Object)id (String)Product IDname (String)Product namedescription (Array)Textual descriptiontraffic (String)Free traffic quotadist (Array)Available distributions@deprecated arch (Array)Available distribution architectureslang (Array)Available distribution languageslocation (Array)Available locationsprices (Array)(Object)location (String)Locationprice (Object)

|                       |                                                                                 |
| --------------------- | ------------------------------------------------------------------------------- |
| net (String)          | Monthly price in euros                                                          |
| gross (String)        | Monthly price in euros with VAT                                                 |
| hourly_net (String)   | Hourly price in euros, if the product is billed hourly, null otherwise          |
| hourly_gross (String) | Hourly price in euros with VAT, if the product is billed hourly, null otherwise |

price_setup (Object)

|                |                                |
| -------------- | ------------------------------ |
| net (String)   | One time fee in euros          |
| gross (String) | One time fee in euros with VAT |

orderable_addons (Array)(Object)id (String)Addon IDname (String)Addon namelocation (String)Locationmin (Integer)Minimum orderable amountmax (Integer)Maximum orderable amountprices (Array)(Object)location (String)Locationprice (Object)

|                       |                                                                                 |
| --------------------- | ------------------------------------------------------------------------------- |
| net (String)          | Monthly price in euros                                                          |
| gross (String)        | Monthly price in euros with VAT                                                 |
| hourly_net (String)   | Hourly price in euros, if the product is billed hourly, null otherwise          |
| hourly_gross (String) | Hourly price in euros with VAT, if the product is billed hourly, null otherwise |

price_setup (Object)

|                |                                |
| -------------- | ------------------------------ |
| net (String)   | One time fee in euros          |
| gross (String) | One time fee in euros with VAT |

### Errors

| Status | Code      | Description       |
| ------ | --------- | ----------------- |
| 404    | NOT_FOUND | Product not found |

## GET /order/server/transaction

    curl -u "user:password" https://robot-ws.your-server.de/order/server/transaction


    [\
      {\
        "transaction":{\
          "id":"B20150121-344957-251478",\
          "date":"2015-01-21T12:30:43+01:00",\
          "status":"in process",\
          "server_number":null,\
          "server_ip":null,\
          "authorized_key":[\
    \
          ],\
          "host_key":[\
    \
          ],\
          "comment":null,\
          "product":{\
            "id":"VX6",\
            "name":"vServer VX6",\
            "description":[\
              "Single-Core CPU",\
              "1 GB RAM",\
              "25 GB HDD",\
              "No telephone support"\
            ],\
            "traffic":"2 TB",\
            "dist":"Rescue system",\
            "@deprecated arch":"64",\
            "lang":"en",\
            "location":null\
          },\
          "addons":[\
            "primary_ipv4"\
          ]\
        }\
      },\
      {\
        "transaction":{\
          "id":"B20150121-344958-251479",\
          "date":"2015-01-21T12:54:01+01:00",\
          "status":"ready",\
          "server_number":107239,\
          "server_ip":"188.40.1.1",\
          "authorized_key":[\
            {\
              "key":{\
                "name":"key1",\
                "fingerprint":"15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb",\
                "type":"ED25519",\
                "size":256\
              }\
            }\
          ],\
          "host_key":[\
            {\
              "key":{\
                "fingerprint":"c1:e4:08:73:dd:f7:e9:d1:94:ab:e9:0f:28:b2:d2:ed",\
                "type":"DSA",\
                "size":1024\
              }\
            }\
          ],\
          "comment":null,\
          "product":{\
            "id":"EX40",\
            "name":"Dedicated Server EX40",\
            "description":[\
              "Intel\u00ae Core\u2122 i7-4770 Quad-Core Haswell",\
              "32 GB DDR3 RAM",\
              "2 x 2 TB SATA 6 Gb\/s Enterprise HDD; 7200 rpm(Software-RAID 1)",\
              "1 Gbit\/s bandwidth"\
            ],\
            "traffic":"30 TB",\
            "dist":"Debian 7.7 minimal",\
            "@deprecated arch":"64",\
            "lang":"en",\
            "location":"FSN1"\
          },\
          "addons":[\
    \
          ]\
        }\
      }\
    ]

### Description

Overview of all server orders within the last 30 days

### Request limit

500 requests per 1 hour

### Output

(Array)transaction (Object)id (String)Transaction IDdate (String)Transaction datestatus (String)Transaction status, "ready", "in process" or "cancelled"server_number (Integer)Server ID if transaction status is "ready", null otherwiseserver_ip (String)Server main IP address if transaction status is "ready", null otherwiseauthorized_key (Array)Array with supplied public SSH keyshost_key (Array)Array with servers public host keyscomment (String)Supplied order commentproduct (Object)

|                            |                                   |
| -------------------------- | --------------------------------- |
| id (String)                | Product ID                        |
| name (String)              | Product name                      |
| description (Array)        | Textual description               |
| traffic (String)           | Free traffic quota                |
| dist (String)              | Ordered distribution              |
| @deprecated arch (Integer) | Ordered distribution architecture |
| lang (String)              | Ordered distribution language     |
| location (String)          | Ordered location                  |

addons (Array)

|          |          |
| -------- | -------- |
| (String) | Addon ID |

### Errors

| Status | Code      | Description           |
| ------ | --------- | --------------------- |
| 404    | NOT_FOUND | No transactions found |

## POST /order/server/transaction

> Order a IPv6-only server

    curl -u "user:password" https://robot-ws.your-server.de/order/server/transaction \
    --data-urlencode 'product_id=EX40' \
    --data-urlencode 'dist=Debian 7.7 minimal' \
    --data-urlencode 'authorized_key[]=15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb'


    {
      "transaction":{
        "id":"B20150121-344958-251479",
        "date":"2015-01-21T12:54:01+01:00",
        "status":"in process",
        "server_number":null,
        "server_ip":null,
        "authorized_key":[\
          {\
            "key":{\
              "name":"key1",\
              "fingerprint":"15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb",\
              "type":"ED25519",\
              "size":256\
            }\
          }\
        ],
        "host_key":[\
    \
        ],
        "comment":null,
        "product":{
          "id":"EX40",
          "name":"Dedicated Server EX40",
          "description":[\
            "Intel\u00ae Core\u2122 i7-4770 Quad-Core Haswell",\
            "32 GB DDR3 RAM",\
            "2 x 2 TB SATA 6 Gb\/s Enterprise HDD; 7200 rpm(Software-RAID 1)",\
            "1 Gbit\/s bandwidth"\
          ],
          "traffic":"30 TB",
          "dist":"Debian 7.7 minimal",
          "@deprecated arch":"64",
          "lang":"en",
          "location":"FSN1"
        },
        "addons":[\
    \
        ]
      }
    }

> Order a server with Primary IPv4 addon

    curl -u "user:password" https://robot-ws.your-server.de/order/server/transaction \
    --data-urlencode 'product_id=EX40' \
    --data-urlencode 'dist=Debian 7.7 minimal' \
    --data-urlencode 'authorized_key[]=15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb' \
    --data-urlencode 'addon[]=primary_ipv4'


    {
      "transaction":{
        "id":"B20150121-344958-251479",
        "date":"2015-01-21T12:54:01+01:00",
        "status":"in process",
        "server_number":null,
        "server_ip":null,
        "authorized_key":[\
          {\
            "key":{\
              "name":"key1",\
              "fingerprint":"15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb",\
              "type":"ED25519",\
              "size":256\
            }\
          }\
        ],
        "host_key":[\
    \
        ],
        "comment":null,
        "product":{
          "id":"EX40",
          "name":"Dedicated Server EX40",
          "description":[\
            "Intel\u00ae Core\u2122 i7-4770 Quad-Core Haswell",\
            "32 GB DDR3 RAM",\
            "2 x 2 TB SATA 6 Gb\/s Enterprise HDD; 7200 rpm(Software-RAID 1)",\
            "1 Gbit\/s bandwidth"\
          ],
          "traffic":"30 TB",
          "dist":"Debian 7.7 minimal",
          "@deprecated arch":"64",
          "lang":"en",
          "location":"FSN1"
        },
        "addons":[\
          "primary_ipv4"\
        ]
      }
    }

### Description

Order a new server. If the order is successful, the status code 201 CREATED is returned.

### Request limit

20 requests per day

### Input

| Name               | Description                                                                                                        |
| ------------------ | ------------------------------------------------------------------------------------------------------------------ |
| product_id         | Product ID                                                                                                         |
| authorized_key\[\] | One or more SSH key fingerprints (Optional, you can use either parameter "authorized_key" or parameter "password") |
| password           | Root password (Optional: you can use either parameter "authorized_key" or parameter "password")                    |
| location           | The desired location                                                                                               |
| dist               | Distribution name which should be preinstalled (optional)                                                          |
| @deprecated arch   | Architecture of preinstalled distribution (optional)                                                               |
| lang               | Language of preinstalled distribution (optional)                                                                   |
| comment            | Order comment (optional); Please note that if a comment is supplied, the order will be processed manually.         |
| addon\[\]          | Array of addon IDs (optional)                                                                                      |
| test               | The order will not be processed if set to "true" (optional)                                                        |

The parameter "addon" is optional. If you do not specify the parameter, the server will be ordered without an IPv4 address by default. If you want to order a server with an IPv4 address, you can supply the value "primary_ipv4".

### Output

transaction (Object)id (String)Transaction IDdate (String)Transaction datestatus (String)Transaction status, "ready", "in process" or "cancelled"server_number (Integer)Server ID if transaction status is "ready", null otherwiseserver_ip (String)Server main IP address if transaction status is "ready", null otherwiseauthorized_key (Array)Array with supplied public SSH keyshost_key (Array)Array with servers public host keyscomment (String)Supplied order commentproduct (Object)

|                            |                                   |
| -------------------------- | --------------------------------- |
| id (String)                | Product ID                        |
| name (String)              | Product name                      |
| description (Array)        | Textual description               |
| traffic (String)           | Free traffic quota                |
| dist (String)              | Ordered distribution              |
| @deprecated arch (Integer) | Ordered distribution architecture |
| lang (String)              | Ordered distribution language     |
| location (String)          | Ordered location                  |

addons (Array)

|          |          |
| -------- | -------- |
| (String) | Addon ID |

### Errors

| Status | Code           | Description                                     |
| ------ | -------------- | ----------------------------------------------- |
| 400    | INVALID_INPUT  | Invalid input parameters                        |
| 500    | INTERNAL_ERROR | The transaction failed due to an internal error |

## GET /order/server/transaction/{id}

    curl -u "user:password" https://robot-ws.your-server.de/order/server/transaction/B20150121-344958-251479


    {
      "transaction":{
        "id":"B20150121-344958-251479",
        "date":"2015-01-21T12:54:01+01:00",
        "status":"ready",
        "server_number":107239,
        "server_ip":"188.40.1.1",
        "authorized_key":[\
          {\
            "key":{\
              "name":"key1",\
              "fingerprint":"15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb",\
              "type":"ED25519",\
              "size":256\
            }\
          }\
        ],
        "host_key":[\
          {\
            "key":{\
              "fingerprint":"c1:e4:08:73:dd:f7:e9:d1:94:ab:e9:0f:28:b2:d2:ed",\
              "type":"DSA",\
              "size":1024\
            }\
          }\
        ],
        "comment":null,
        "product":{
          "id":"EX40",
          "name":"Dedicated Server EX40",
          "description":[\
            "Intel\u00ae Core\u2122 i7-4770 Quad-Core Haswell",\
            "32 GB DDR3 RAM",\
            "2 x 2 TB SATA 6 Gb\/s Enterprise HDD; 7200 rpm(Software-RAID 1)",\
            "1 Gbit\/s bandwidth"\
          ],
          "traffic":"30 TB",
          "dist":"Debian 7.7 minimal",
          "@deprecated arch":"64",
          "lang":"en",
          "location":"FSN1"
        },
        "addons":[\
          "primary_ipv4"\
        ]
      }
    }

### Description

Query a specific order transaction

### Request limit

500 requests per 1 hour

### Output

transaction (Object)id (String)Transaction IDdate (String)Transaction datestatus (String)Transaction status, "ready", "in process" or "cancelled"server_number (Integer)Server ID if transaction status is "ready", null otherwiseserver_ip (String)Server main IP address if transaction status is "ready", null otherwiseauthorized_key (Array)Array with supplied public SSH keyshost_key (Array)Array with servers public host keyscomment (String)Supplied order commentproduct (Object)

|                            |                                   |
| -------------------------- | --------------------------------- |
| id (String)                | Product ID                        |
| name (String)              | Product name                      |
| description (Array)        | Textual description               |
| traffic (String)           | Free traffic quota                |
| dist (String)              | Ordered distribution              |
| @deprecated arch (Integer) | Ordered distribution architecture |
| lang (String)              | Ordered distribution language     |
| location (String)          | Ordered location                  |

addons (Array)

|          |          |
| -------- | -------- |
| (String) | Addon ID |

### Errors

| Status | Code      | Description           |
| ------ | --------- | --------------------- |
| 404    | NOT_FOUND | Transaction not found |

## GET /order/server_market/product

    curl -u "user:password" https://robot-ws.your-server.de/order/server_market/product


    [\
      {\
        "product":{\
          "id":276112,\
          "name":"SB34",\
          "description":[\
            "AMD Athlon 64 6000+ X2",\
            "4x RAM 2048 MB DDR2",\
            "2x HDD 750 GB SATA",\
            "RAID Controller 2-Port SATA PCI - 3ware 8006-2LP"\
          ],\
          "traffic":"20 TB",\
          "dist":[\
            "Rescue system"\
          ],\
          "@deprecated arch":[\
            64\
          ],\
          "lang":[\
            "en"\
          ],\
          "cpu":"AMD Athlon 64 6000+ X2",\
          "cpu_benchmark":1580,\
          "memory_size":8,\
          "hdd_size":750,\
          "hdd_text":"ENT.HDD ECC INIC",\
          "hdd_count":2,\
          "datacenter":"NBG1-DC1",\
          "network_speed":"100 Mbit\/s",\
          "price":"28.57",\
          "price_hourly":"0.0458",\
          "price_setup":"0.00",\
          "price_vat":"28.57",\
          "price_hourly_vat":"0.0458",\
          "price_setup_vat":"0.00",\
          "fixed_price":false,\
          "next_reduce":-87634,\
          "next_reduce_date":"2018-05-01 12:22:00",\
          "orderable_addons":[\
            {\
              "id":"primary_ipv4",\
              "name":"Primary IPv4",\
              "min":0,\
              "max":1,\
              "prices":[\
                {\
                  "location":"FSN1",\
                  "price":{\
                    "net":"1.7000",\
                    "gross":"1.7000",\
                    "hourly_net":"0.0027",\
                    "hourly_gross":"0.0027"\
                  },\
                  "price_setup":{\
                    "net":"0.0000",\
                    "gross":"0.0000"\
                  }\
                },\
                {\
                  "location":"NBG1",\
                  "price":{\
                    "net":"1.7000",\
                    "gross":"1.7000",\
                    "hourly_net":"0.0027",\
                    "hourly_gross":"0.0027"\
                  },\
                  "price_setup":{\
                    "net":"0.0000",\
                    "gross":"0.0000"\
                  }\
                }\
              ]\
            }\
          ]\
        }\
      },\
      {\
        "product":{\
          "id":282323,\
          "name":"SB109",\
          "description":[\
            "Intel Core i7 980x",\
            "6x RAM 4096 MB DDR3",\
            "2x SSD 120 GB SATA",\
            "NIC 1000Mbit PCI - Intel Pro1000GT PWLA8391GT",\
            "RAID Controller 4-Port SATA PCI-E - Adaptec 5405"\
          ],\
          "traffic":"20 TB",\
          "dist":[\
            "Rescue system"\
          ],\
          "@deprecated arch":[\
            64\
          ],\
          "lang":[\
            "en"\
          ],\
          "cpu":"Intel Core i7 980x",\
          "cpu_benchmark":8944,\
          "memory_size":24,\
          "hdd_size":120,\
          "hdd_text":"ESAS HWR",\
          "hdd_count":2,\
          "datacenter":"FSN1-DC4",\
          "network_speed":"200 Mbit\/s",\
          "price":"91.60",\
          "price_hourly":"0.1468",\
          "price_setup":"0.00",\
          "price_vat":"91.60",\
          "price_hourly_vat":"0.1468",\
          "price_setup_vat":"0.00",\
          "fixed_price":false,\
          "next_reduce":-10800,\
          "next_reduce_date":"2018-05-01 12:22:00",\
          "orderable_addons":[\
            {\
              "id":"primary_ipv4",\
              "name":"Primary IPv4",\
              "min":0,\
              "max":1,\
              "prices":[\
                {\
                  "location":"FSN1",\
                  "price":{\
                    "net":"1.7000",\
                    "gross":"1.7000",\
                    "hourly_net":"0.0027",\
                    "hourly_gross":"0.0027"\
                  },\
                  "price_setup":{\
                    "net":"0.0000",\
                    "gross":"0.0000"\
                  }\
                },\
                {\
                  "location":"NBG1",\
                  "price":{\
                    "net":"1.7000",\
                    "gross":"1.7000",\
                    "hourly_net":"0.0027",\
                    "hourly_gross":"0.0027"\
                  },\
                  "price_setup":{\
                    "net":"0.0000",\
                    "gross":"0.0000"\
                  }\
                }\
              ]\
            }\
          ]\
        }\
      }\
    ]

### Description

Product overview of currently offered server market products

### Request limit

500 requests per 1 hour

### Output

(Array)product (Object)id (Integer)Product IDname (String)Product namedescription (Array)Textual descriptiontraffic (String)Free traffic quotadist (Array)Available distributions@deprecated arch (Array)Available distribution architectureslang (Array)Available distribution languagescpu (String)CPU model namecpu_benchmark (Integer)CPU benchmark valuememory_size (Integer)Main memory size in GBhdd_size (Integer)Drive size in GBhdd_text (String)Drive special tagshdd_count (Integer)Drive countdatacenter (String)Data centernetwork_speed (String)Server network speedprice (String)Monthly price in eurosprice_hourly (String)Hourly price in euros, if the product is billed hourly, null otherwiseprice_setup (String)One time fee in eurosprice_vat (String)Monthly price in euros with VATprice_hourly_vat (String)Hourly price in euros with VAT, if the product is billed hourly, null otherwiseprice_setup_vat (String)One time fee in euros with VATfixed_price (Boolean)Set to "true" if product has a fixed pricenext_reduce (Integer)Countdown until next price reduction in secondsnext_reduce_date (String)Next price reduction dateorderable_addons (Array)(Object)id (String)Addon IDname (String)Addon namelocation (String)Locationmin (Integer)Minimum orderable amountmax (Integer)Maximum orderable amountprices (Array)(Object)location (String)Locationprice (Object)

|                       |                                                                                 |
| --------------------- | ------------------------------------------------------------------------------- |
| net (String)          | Monthly price in euros                                                          |
| gross (String)        | Monthly price in euros with VAT                                                 |
| hourly_net (String)   | Hourly price in euros, if the product is billed hourly, null otherwise          |
| hourly_gross (String) | Hourly price in euros with VAT, if the product is billed hourly, null otherwise |

price_setup (Object)

|                |                                |
| -------------- | ------------------------------ |
| net (String)   | One time fee in euros          |
| gross (String) | One time fee in euros with VAT |

### Errors

| Status | Code      | Description       |
| ------ | --------- | ----------------- |
| 404    | NOT_FOUND | No products found |

## GET /order/server_market/product/{product-id}

    curl -u "user:password" https://robot-ws.your-server.de/order/server_market/product/282323


    {
      "product":{
        "id":282323,
        "name":"SB109",
        "description":[\
          "Intel Core i7 980x",\
          "6x RAM 4096 MB DDR3",\
          "2x SSD 120 GB SATA",\
          "NIC 1000Mbit PCI - Intel Pro1000GT PWLA8391GT",\
          "RAID Controller 4-Port SATA PCI-E - Adaptec 5405"\
        ],
        "traffic":"20 TB",
        "dist":[\
          "Rescue system"\
        ],
        "@deprecated arch":[\
          64\
        ],
        "lang":[\
          "en"\
        ],
        "cpu":"Intel Core i7 980x",
        "cpu_benchmark":8944,
        "memory_size":24,
        "hdd_size":120,
        "hdd_text":"ENT.HDD ECC INIC",
        "hdd_count":2,
        "datacenter":"FSN1-DC4",
        "network_speed":"100 Mbit\/s",
        "price":"91.60",
        "price_hourly":"0.1468",
        "price_setup":"0.00",
        "price_vat":"91.60",
        "price_hourly_vat":"0.1468",
        "price_setup_vat":"0.00",
        "fixed_price":false,
        "next_reduce":-10800,
        "next_reduce_date":"2018-05-01 12:22:00",
        "orderable_addons":[\
          {\
            "id":"primary_ipv4",\
            "name":"Primary IPv4",\
            "min":0,\
            "max":1,\
            "prices":[\
              {\
                "location":"FSN1",\
                "price":{\
                  "net":"1.7000",\
                  "gross":"1.7000",\
                  "hourly_net":"0.0027",\
                  "hourly_gross":"0.0027"\
                },\
                "price_setup":{\
                  "net":"0.0000",\
                  "gross":"0.0000"\
                }\
              },\
              {\
                "location":"NBG1",\
                "price":{\
                  "net":"1.7000",\
                  "gross":"1.7000",\
                  "hourly_net":"0.0027",\
                  "hourly_gross":"0.0027"\
                },\
                "price_setup":{\
                  "net":"0.0000",\
                  "gross":"0.0000"\
                }\
              }\
            ]\
          }\
        ]
      }
    }

### Description

Query a specific server market product

### Request limit

500 requests per 1 hour

### Output

product (Object)id (Integer)Product IDname (String)Product namedescription (Array)Textual descriptiontraffic (String)Free traffic quotadist (Array)Available distributions@deprecated arch (Array)Available distribution architectureslang (Array)Available distribution languagescpu (String)CPU model namecpu_benchmark (Integer)CPU benchmark valuememory_size (Integer)Main memory size in GBhdd_size (Integer)Drive size in GBhdd_text (String)Drive special tagshdd_count (Integer)Drive countdatacenter (String)Data centernetwork_speed (String)Server network speedprice (String)Monthly price in eurosprice_hourly (String)Hourly price in euros, if the product is billed hourly, null otherwiseprice_setup (String)One time fee in eurosprice_vat (String)Monthly price in euros with VATprice_hourly_vat (String)Hourly price in euros with VAT, if the product is billed hourly, null otherwiseprice_setup_vat (String)One time fee in euros with VATfixed_price (Boolean)Set to "true" if product has a fixed pricenext_reduce (Integer)Countdown until next price reduction in secondsnext_reduce_date (String)Next price reduction dateorderable_addons (Array)(Object)id (String)Addon IDname (String)Addon namelocation (String)Locationmin (Integer)Minimum orderable amountmax (Integer)Maximum orderable amountprices (Array)(Object)location (String)Locationprice (Object)

|                       |                                                                                 |
| --------------------- | ------------------------------------------------------------------------------- |
| net (String)          | Monthly price in euros                                                          |
| gross (String)        | Monthly price in euros with VAT                                                 |
| hourly_net (String)   | Hourly price in euros, if the product is billed hourly, null otherwise          |
| hourly_gross (String) | Hourly price in euros with VAT, if the product is billed hourly, null otherwise |

price_setup (Object)

|                |                                |
| -------------- | ------------------------------ |
| net (String)   | One time fee in euros          |
| gross (String) | One time fee in euros with VAT |

### Errors

| Status | Code      | Description       |
| ------ | --------- | ----------------- |
| 404    | NOT_FOUND | Product not found |

## GET /order/server_market/transaction

    curl -u "user:password" https://robot-ws.your-server.de/order/server_market/transaction


    [\
      {\
        "transaction":{\
          "id":"B20150121-344957-251478",\
          "date":"2015-01-21T12:30:43+01:00",\
          "status":"in process",\
          "server_number":null,\
          "server_ip":null,\
          "authorized_key":[\
    \
          ],\
          "host_key":[\
    \
          ],\
          "comment":null,\
          "product":{\
            "id":283693,\
            "name":"SB110",\
            "description":[\
              "Intel Core i7 980x",\
              "6x RAM 4096 MB DDR3",\
              "2x HDD 1,5 TB SATA",\
              "2x SSD 120 GB SATA"\
            ],\
            "traffic":"20 TB",\
            "dist":"Rescue system",\
            "@deprecated arch":"64",\
            "lang":"en",\
            "cpu":"Intel Core i7 980x",\
            "cpu_benchmark":8944,\
            "memory_size":24,\
            "hdd_size":1536,\
            "hdd_text":"ENT.HDD ECC INIC",\
            "hdd_count":2,\
            "datacenter":"FSN1-DC5",\
            "network_speed":"100 Mbit\/s",\
            "fixed_price":true,\
            "next_reduce":0,\
            "next_reduce_date":"2018-05-01 12:22:00"\
          }\
        }\
      },\
      {\
        "transaction":{\
          "id":"B20150121-344958-251479",\
          "date":"2015-01-21T12:54:01+01:00",\
          "status":"ready",\
          "server_number":107239,\
          "server_ip":"188.40.1.1",\
          "authorized_key":[\
            {\
              "key":{\
                "name":"key1",\
                "fingerprint":"15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb",\
                "type":"ED25519",\
                "size":256\
              }\
            }\
          ],\
          "host_key":[\
            {\
              "key":{\
                "fingerprint":"c1:e4:08:73:dd:f7:e9:d1:94:ab:e9:0f:28:b2:d2:ed",\
                "type":"DSA",\
                "size":1024\
              }\
            }\
          ],\
          "comment":null,\
          "product":{\
            "id":277254,\
            "name":"SB114",\
            "description":[\
              "Intel Core i7 950",\
              "6x RAM 2048 MB DDR3",\
              "7x HDD 1,5 TB SATA"\
            ],\
            "traffic":"20 TB",\
            "dist":"Rescue system",\
            "@deprecated arch":"64",\
            "lang":"en",\
            "cpu":"Intel Core i7 950",\
            "cpu_benchmark":5682,\
            "memory_size":12,\
            "hdd_size":1536,\
            "hdd_text":"ENT.HDD ECC INIC",\
            "hdd_count":7,\
            "datacenter":"FSN1-DC5",\
            "network_speed":"100 Mbit\/s",\
            "fixed_price":true,\
            "next_reduce":0,\
            "next_reduce_date":"2018-05-01 12:22:00"\
          }\
        }\
      }\
    ]

### Description

Overview of all server orders within the last 30 days

### Request limit

500 requests per 1 hour

### Output

(Array)transaction (Object)id (String)Transaction IDdate (String)Transaction datestatus (String)Transaction status, "ready", "in process" or "cancelled"server_number (Integer)Server ID if transaction status is "ready", null otherwiseserver_ip (String)Server main IP address if transaction status is "ready", null otherwiseauthorized_key (Array)Array with supplied public SSH keyshost_key (Array)Array with servers public host keyscomment (String)Supplied order commentproduct (Object)

|                           |                                   |
| ------------------------- | --------------------------------- |
| id (Integer)              | Product ID                        |
| name (String)             | Product name                      |
| description (Array)       | Textual description               |
| traffic (String)          | Free traffic quota                |
| dist (String)             | Ordered distribution              |
| @deprecated arch (String) | Ordered distribution architecture |
| lang (String)             | Ordered distribution language     |
| cpu (String)              | CPU model name                    |
| cpu_benchmark (Integer)   | CPU benchmark value               |
| memory_size (Integer)     | Main memory size in GB            |
| hdd_size (Integer)        | Drive size in GB                  |
| hdd_text (String)         | Drive special tags                |
| hdd_count (Integer)       | Drive count                       |
| datacenter (String)       | Data center                       |
| network_speed (String)    | Server network speed              |
| fixed_price (Boolean)     | true                              |
| next_reduce (Integer)     | 0                                 |
| next_reduce_date (String) | 2018-05-01 12:22:00               |

### Errors

| Status | Code      | Description           |
| ------ | --------- | --------------------- |
| 404    | NOT_FOUND | No transactions found |

## POST /order/server_market/transaction

> Order a IPv6-only server

    curl -u "user:password" https://robot-ws.your-server.de/order/server_market/transaction \
    --data-urlencode 'product_id=283693' \
    --data-urlencode 'authorized_key[]=15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb'


    {
      "transaction":{
        "id":"B20150121-344958-251479",
        "date":"2015-01-21T12:54:01+01:00",
        "status":"in process",
        "server_number":null,
        "server_ip":null,
        "authorized_key":[\
          {\
            "key":{\
              "name":"key1",\
              "fingerprint":"15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb",\
              "type":"ED25519",\
              "size":256\
            }\
          }\
        ],
        "host_key":[\
    \
        ],
        "comment":null,
        "product":{
          "id":283693,
          "name":"SB110",
          "description":[\
            "Intel Core i7 980x",\
            "6x RAM 4096 MB DDR3",\
            "2x HDD 1,5 TB SATA",\
            "2x SSD 120 GB SATA"\
          ],
          "traffic":"20 TB",
          "dist":"Rescue system",
          "@deprecated arch":"64",
          "lang":"en",
          "cpu":"Intel Core i7 980x",
          "cpu_benchmark":8944,
          "memory_size":24,
          "hdd_size":1536,
          "hdd_text":"ENT.HDD ECC INIC",
          "hdd_count":2,
          "datacenter":"FSN1-DC5",
          "network_speed":"100 Mbit\/s"
        },
        "addons":[\
    \
        ]
      }
    }

> Order a server with Primary IPv4 addon

    curl -u "user:password" https://robot-ws.your-server.de/order/server_market/transaction \
    --data-urlencode 'product_id=283693' \
    --data-urlencode 'authorized_key[]=15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb' \
    --data-urlencode 'addon[]=primary_ipv4'


    {
      "transaction":{
        "id":"B20150121-344958-251479",
        "date":"2015-01-21T12:54:01+01:00",
        "status":"in process",
        "server_number":null,
        "server_ip":null,
        "authorized_key":[\
          {\
            "key":{\
              "name":"key1",\
              "fingerprint":"15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb",\
              "type":"ED25519",\
              "size":256\
            }\
          }\
        ],
        "host_key":[\
    \
        ],
        "comment":null,
        "product":{
          "id":283693,
          "name":"SB110",
          "description":[\
            "Intel Core i7 980x",\
            "6x RAM 4096 MB DDR3",\
            "2x HDD 1,5 TB SATA",\
            "2x SSD 120 GB SATA"\
          ],
          "traffic":"20 TB",
          "dist":"Rescue system",
          "@deprecated arch":"64",
          "lang":"en",
          "cpu":"Intel Core i7 980x",
          "cpu_benchmark":8944,
          "memory_size":24,
          "hdd_size":1536,
          "hdd_text":"ENT.HDD ECC INIC",
          "hdd_count":2,
          "datacenter":"FSN1-DC5",
          "network_speed":"100 Mbit\/s"
        },
        "addons":[\
          "primary_ipv4"\
        ]
      }
    }

### Description

Order a new server from the server market. If the order is successful, the status code 201 CREATED is returned.

### Request limit

20 requests per day

### Input

| Name               | Description                                                                                                        |
| ------------------ | ------------------------------------------------------------------------------------------------------------------ |
| product_id         | Product ID                                                                                                         |
| authorized_key\[\] | One or more SSH key fingerprints (Optional, you can use either parameter "authorized_key" or parameter "password") |
| password           | Root password (Optional: you can use either parameter "authorized_key" or parameter "password")                    |
| dist               | Distribution name which should be preinstalled (optional)                                                          |
| @deprecated arch   | Architecture of preinstalled distribution (optional)                                                               |
| lang               | Language of preinstalled distribution (optional)                                                                   |
| comment            | Order comment (optional); Please note that if a comment is supplied, the order will be processed manually.         |
| addon\[\]          | Array of addon IDs (optional)                                                                                      |
| test               | The order will not be processed if set to "true" (optional)                                                        |

The parameter "addon" is optional. If you do not specify the parameter, the server will be ordered without an IPv4 address by default. If you want to order a server with an IPv4 address, you can supply the value "primary_ipv4".

### Output

transaction (Object)id (String)Transaction IDdate (String)Transaction datestatus (String)Transaction status, "ready", "in process" or "cancelled"server_number (Integer)Server ID if transaction status is "ready", null otherwiseserver_ip (String)Server main IP address if transaction status is "ready", null otherwiseauthorized_key (Array)Array with supplied public SSH keyshost_key (Array)Array with servers public host keyscomment (String)Supplied order commentproduct (Object)

|                           |                                   |
| ------------------------- | --------------------------------- |
| id (Integer)              | Product ID                        |
| name (String)             | Product name                      |
| description (Array)       | Textual description               |
| traffic (String)          | Free traffic quota                |
| dist (String)             | Ordered distribution              |
| @deprecated arch (String) | Ordered distribution architecture |
| lang (String)             | Ordered distribution language     |
| cpu (String)              | CPU model name                    |
| cpu_benchmark (Integer)   | CPU benchmark value               |
| memory_size (Integer)     | Main memory size in GB            |
| hdd_size (Integer)        | Drive size in GB                  |
| hdd_text (String)         | Drive special tags                |
| hdd_count (Integer)       | Drive count                       |
| datacenter (String)       | Data center                       |
| network_speed (String)    | Server network speed              |

addons (Array)

|          |          |
| -------- | -------- |
| (String) | Addon ID |

### Errors

| Status | Code           | Description                                     |
| ------ | -------------- | ----------------------------------------------- |
| 400    | INVALID_INPUT  | Invalid input parameters                        |
| 500    | INTERNAL_ERROR | The transaction failed due to an internal error |

## GET /order/server_market/transaction/{id}

    curl -u "user:password" https://robot-ws.your-server.de/order/server_market/transaction/B20150121-344958-251479


    {
      "transaction":{
        "id":"B20150121-344958-251479",
        "date":"2015-01-21T12:54:01+01:00",
        "status":"in process",
        "server_number":null,
        "server_ip":null,
        "authorized_key":[\
          {\
            "key":{\
              "name":"key1",\
              "fingerprint":"15:28:b0:03:95:f0:77:b3:10:56:15:6b:77:22:a5:bb",\
              "type":"ED25519",\
              "size":256\
            }\
          }\
        ],
        "host_key":[\
    \
        ],
        "comment":null,
        "product":{
          "id":283693,
          "name":"SB110",
          "description":[\
            "Intel Core i7 980x",\
            "6x RAM 4096 MB DDR3",\
            "2x HDD 1,5 TB SATA",\
            "2x SSD 120 GB SATA"\
          ],
          "traffic":"20 TB",
          "dist":"Rescue system",
          "@deprecated arch":"64",
          "lang":"en",
          "cpu":"Intel Core i7 980x",
          "cpu_benchmark":8944,
          "memory_size":24,
          "hdd_size":1536,
          "hdd_text":"ENT.HDD ECC INIC",
          "hdd_count":2,
          "datacenter":"FSN1-DC5",
          "network_speed":"100 Mbit\/s"
        },
        "addons":[\
          "primary_ipv4"\
        ]
      }
    }

### Description

Query a specific order transaction

### Request limit

500 requests per 1 hour

### Output

transaction (Object)id (String)Transaction IDdate (String)Transaction datestatus (String)Transaction status, "ready", "in process" or "cancelled"server_number (Integer)Server ID if transaction status is "ready", null otherwiseserver_ip (String)Server main IP address if transaction status is "ready", null otherwiseauthorized_key (Array)Array with supplied public SSH keyshost_key (Array)Array with servers public host keyscomment (String)Supplied order commentproduct (Object)

|                           |                                   |
| ------------------------- | --------------------------------- |
| id (Integer)              | Product ID                        |
| name (String)             | Product name                      |
| description (Array)       | Textual description               |
| traffic (String)          | Free traffic quota                |
| dist (String)             | Ordered distribution              |
| @deprecated arch (String) | Ordered distribution architecture |
| lang (String)             | Ordered distribution language     |
| cpu (String)              | CPU model name                    |
| cpu_benchmark (Integer)   | CPU benchmark value               |
| memory_size (Integer)     | Main memory size in GB            |
| hdd_size (Integer)        | Drive size in GB                  |
| hdd_text (String)         | Drive special tags                |
| hdd_count (Integer)       | Drive count                       |
| datacenter (String)       | Data center                       |
| network_speed (String)    | Server network speed              |

addons (Array)

|          |          |
| -------- | -------- |
| (String) | Addon ID |

### Errors

| Status | Code      | Description           |
| ------ | --------- | --------------------- |
| 404    | NOT_FOUND | Transaction not found |

## GET /order/server_addon/{server-number}/product

    curl -u "user:password" https://robot-ws.your-server.de/order/server_addon/123/product


    [\
      {\
        "product":{\
          "id":"additional_ipv4",\
          "name":"Additional IP address",\
          "type":"ip_ipv4",\
          "price":{\
            "location":"NBG1",\
            "price":{\
              "net":"0.8403",\
              "gross":"0.8403",\
              "hourly_net":"0.0014",\
              "hourly_gross":"0.0014"\
            },\
            "price_setup":{\
              "net":"19.0000",\
              "gross":"19.0000"\
            }\
          }\
        }\
      },\
      {\
        "product":{\
          "id":"subnet_ipv4_29",\
          "name":"Additional subnet \/29 (monthly charge)",\
          "type":"subnet_ipv4",\
          "price":{\
            "location":"NBG1",\
            "price":{\
              "net":"6.7227",\
              "gross":"6.7227",\
              "hourly_net":"0.0108",\
              "hourly_gross":"0.0108"\
            },\
            "price_setup":{\
              "net":"152.0000",\
              "gross":"152.0000"\
            }\
          }\
        }\
      }\
    ]

### Description

Product overview of available server addons for a server

### Request limit

500 requests per 1 hour

### Output

(Array)product (Object)id (String)Product IDname (String)Product nametype (String)Product typeprice (Object)location (String)Locationprice (Object)

|                       |                                                                                 |
| --------------------- | ------------------------------------------------------------------------------- |
| net (String)          | Monthly price in euros                                                          |
| gross (String)        | Monthly price in euros with VAT                                                 |
| hourly_net (String)   | Hourly price in euros, if the product is billed hourly, null otherwise          |
| hourly_gross (String) | Hourly price in euros with VAT, if the product is billed hourly, null otherwise |

price_setup (Object)

|                |                                |
| -------------- | ------------------------------ |
| net (String)   | One time fee in euros          |
| gross (String) | One time fee in euros with VAT |

## GET /order/server_addon/transaction

    curl -u "user:password" https://robot-ws.your-server.de/order/server_addon/transaction


    [\
      {\
        "transaction":{\
          "id":"B20220210-1843193-S33055",\
          "date":"2022-02-10T12:20:11+01:00",\
          "status":"in process",\
          "server_number":123,\
          "product":{\
            "id":"failover_subnet_ipv4_29",\
            "name":"Failover subnet \/29",\
            "price":{\
              "location":"NBG1",\
              "price":{\
                "net":"15.1261",\
                "gross":"15.1261",\
                "hourly_net":"0.0242",\
                "hourly_gross":"0.0242"\
              },\
              "price_setup":{\
                "net":"152.0000",\
                "gross":"152.0000"\
              }\
            }\
          },\
          "resources":[\
    \
          ]\
        }\
      },\
      {\
        "transaction":{\
          "id":"B20220210-1843192-S33051",\
          "date":"2022-02-10T11:20:13+01:00",\
          "status":"ready",\
          "server_number":123,\
          "product":{\
            "id":"failover_subnet_ipv4_29",\
            "name":"Failover subnet \/29",\
            "price":{\
              "location":"NBG1",\
              "price":{\
                "net":"15.1261",\
                "gross":"15.1261",\
                "hourly_net":"0.0242",\
                "hourly_gross":"0.0242"\
              },\
              "price_setup":{\
                "net":"152.0000",\
                "gross":"152.0000"\
              }\
            }\
          },\
          "resources":[\
            {\
              "type":"subnet",\
              "id":"10.0.0.0"\
            }\
          ]\
        }\
      }\
    ]

### Description

Overview of all addon orders within the last 30 days

### Request limit

500 requests per 1 hour

### Output

(Array)transaction (Object)id (String)Transaction IDdate (String)Transaction datestatus (String)Transaction status, "ready", "in process" or "cancelled"server_number (Integer)Server IDproduct (Object)id (String)Product IDname (String)Product nametype (String)Product typeprice (Object)location (String)Locationprice (Object)

|                       |                                                                                 |
| --------------------- | ------------------------------------------------------------------------------- |
| net (String)          | Monthly price in euros                                                          |
| gross (String)        | Monthly price in euros with VAT                                                 |
| hourly_net (String)   | Hourly price in euros, if the product is billed hourly, null otherwise          |
| hourly_gross (String) | Hourly price in euros with VAT, if the product is billed hourly, null otherwise |

price_setup (Object)

|                |                                |
| -------------- | ------------------------------ |
| net (String)   | One time fee in euros          |
| gross (String) | One time fee in euros with VAT |

resources (Array)(Object)

|      |                  |
| ---- | ---------------- |
| type | Type of resource |
| id   | Resource ID      |

## POST /order/server_addon/transaction

> Order a single additional IPv4 address for server 123

    curl -u "user:password" https://robot-ws.your-server.de/order/server_addon/transaction \
    --data-urlencode 'server_number=123' \
    --data-urlencode 'reason=VPS' \
    --data-urlencode 'product_id=additional_ipv4'


    {
      "transaction":{
        "id":"B20220210-1843193-S33055",
        "date":"2022-02-10T12:20:11+01:00",
        "status":"in process",
        "server_number":123,
        "product":{
          "id":"additional_ipv4",
          "name":"Additional IP address",
          "type":"ip_ipv4",
          "price":{
            "location":"NBG1",
            "price":{
              "net":"0.8403",
              "gross":"0.8403",
              "hourly_net":"0.0014",
              "hourly_gross":"0.0014"
            },
            "price_setup":{
              "net":"19.0000",
              "gross":"19.0000"
            }
          }
        },
        "resources":[\
    \
        ]
      }
    }

> Order an IPv4 /29 subnet for server 123 and set routing to additional IP 10.0.0.1

    curl -u "user:password" https://robot-ws.your-server.de/order/server_addon/transaction \
    --data-urlencode 'server_number=123' \
    --data-urlencode 'product_id=additional_ipv4' \
    --data-urlencode 'reason=VPS' \
    --data-urlencode 'gateway=10.0.0.1'


    {
      "transaction":{
        "id":"B20220210-1843193-S33055",
        "date":"2022-02-10T12:20:11+01:00",
        "status":"in process",
        "server_number":123,
        "product":{
          "id":"subnet_ipv4_29",
          "name":"Additional subnet \/29 (monthly charge)",
          "type":"subnet_ipv4",
          "price":{
            "location":"NBG1",
            "price":{
              "net":"6.7227",
              "gross":"6.7227",
              "hourly_net":"0.0108",
              "hourly_gross":"0.0108"
            },
            "price_setup":{
              "net":"152.0000",
              "gross":"152.0000"
            }
          }
        },
        "resources":[\
    \
        ]
      }
    }

### Description

Order an addon for a server. If the order is successful, the status code 201 CREATED will be returned.

### Request limit

20 requests per day

### Input

| Name          | Description                                                                                                            |
| ------------- | ---------------------------------------------------------------------------------------------------------------------- |
| product_id    | Product ID                                                                                                             |
| server_number | Server ID                                                                                                              |
| reason        | RIPE reason: mandatory for addon types "ip_ipv4", "subnet_ipv4" and "failover_subnet_ipv4"                             |
| gateway       | Routing target for subnets: usable for addon type "subnet_ipv4" (Optional: default is the server's primary IP address) |
| test          | The order will not be processed if set to "true" (optional)                                                            |

The parameter "gateway" is optional. If you do not specify the parameter, the subnet that you have ordered will be routed to the server's primary IP address. For addon type "subnet_ipv4" you can use one of the servers additional IPv4 address.

### Output

transaction (Object)id (String)Transaction IDdate (String)Transaction datestatus (String)Transaction status, "ready", "in process" or "cancelled"server_number (Integer)Server IDproduct (Object)id (String)Product IDname (String)Product nametype (String)Product typeprice (Object)location (String)Locationprice (Object)

|                       |                                                                                 |
| --------------------- | ------------------------------------------------------------------------------- |
| net (String)          | Monthly price in euros                                                          |
| gross (String)        | Monthly price in euros with VAT                                                 |
| hourly_net (String)   | Hourly price in euros, if the product is billed hourly, null otherwise          |
| hourly_gross (String) | Hourly price in euros with VAT, if the product is billed hourly, null otherwise |

price_setup (Object)

|                |                                |
| -------------- | ------------------------------ |
| net (String)   | One time fee in euros          |
| gross (String) | One time fee in euros with VAT |

resources (Array)(Object)

|      |                  |
| ---- | ---------------- |
| type | Type of resource |
| id   | Resource ID      |

### Errors

| Status | Code           | Description                                                                          |
| ------ | -------------- | ------------------------------------------------------------------------------------ |
| 400    | INVALID_INPUT  | Invalid input parameters                                                             |
| 409    | CONFLICT       | The transaction cannot be processed due to the reason mentioned in the error message |
| 500    | INTERNAL_ERROR | The transaction failed due to an internal error                                      |

## GET /order/server_addon/transaction/{id}

    curl -u "user:password" https://robot-ws.your-server.de/order/server_addon/transaction/B20220210-1843193-S33055


    {
      "transaction":{
        "id":"B20220210-1843193-S33055",
        "date":"2022-02-10T12:20:11+01:00",
        "status":"in process",
        "server_number":123,
        "product":{
          "id":"failover_subnet_ipv4_29",
          "name":"Failover subnet \/29",
          "price":{
            "location":"NBG1",
            "price":{
              "net":"15.1261",
              "gross":"15.1261",
              "hourly_net":"0.0242",
              "hourly_gross":"0.0242"
            },
            "price_setup":{
              "net":"152.0000",
              "gross":"152.0000"
            }
          }
        },
        "resources":[\
    \
        ]
      }
    }

### Description

Query a specific order transaction

### Request limit

500 requests per 1 hour

### Output

transaction (Object)id (String)Transaction IDdate (String)Transaction datestatus (String)Transaction status, "ready", "in process" or "cancelled"server_number (Integer)Server IDproduct (Object)id (String)Product IDname (String)Product nametype (String)Product typeprice (Object)location (String)Locationprice (Object)

|                       |                                                                                 |
| --------------------- | ------------------------------------------------------------------------------- |
| net (String)          | Monthly price in euros                                                          |
| gross (String)        | Monthly price in euros with VAT                                                 |
| hourly_net (String)   | Hourly price in euros, if the product is billed hourly, null otherwise          |
| hourly_gross (String) | Hourly price in euros with VAT, if the product is billed hourly, null otherwise |

price_setup (Object)

|                |                                |
| -------------- | ------------------------------ |
| net (String)   | One time fee in euros          |
| gross (String) | One time fee in euros with VAT |

resources (Array)(Object)

|               |                  |
| ------------- | ---------------- |
| type (String) | Type of resource |
| id (String)   | Resource ID      |

### Errors

| Status | Code      | Description           |
| ------ | --------- | --------------------- |
| 404    | NOT_FOUND | Transaction not found |

# Storage Box

## GET /storagebox

    curl -u "user:password" https://robot-ws.your-server.de/storagebox


    [\
      {\
        "storagebox":{\
          "id":123456,\
          "login":"u12345",\
          "name":"Backup Server 1",\
          "product":"BX60",\
          "cancelled":false,\
          "locked":false,\
          "location":"FSN1",\
          "linked_server":123456,\
          "paid_until":"2015-10-23"\
        }\
      }\
    ]


    curl -u "user:password" https://robot-ws.your-server.de/storagebox -d linked_server=123456


    {
      "storagebox":{
        "id":123456,
        "login":"u12345",
        "name":"Backup Server 1",
        "product":"BX60",
        "cancelled":false,
        "locked":false,
        "location":"FSN1",
        "linked_server":123456,
        "paid_until":"2015-10-23"
      }
    }

### Description

Query data of all Storage Boxes

### Request limit

200 requests per 1 hour

### Filters

| Name          | Description      |
| ------------- | ---------------- |
| linked_server | Linked Server ID |

### Output

(Array)storagebox (Object)

|                         |                                    |
| ----------------------- | ---------------------------------- |
| id (Integer)            | Storage Box ID                     |
| login (String)          | User name                          |
| name (String)           | Name of the Storage Box            |
| product (String)        | Product name                       |
| cancelled (Boolean)     | Status of Storage Box cancellation |
| locked (Boolean)        | Status of locking                  |
| location (String)       | Location of Storage Box host       |
| linked_server (Integer) | Linked server id                   |
| paid_until (String)     | Paid until date                    |

### Errors

| Status | Code                 | Description            |
| ------ | -------------------- | ---------------------- |
| 404    | STORAGEBOX_NOT_FOUND | No Storage Boxes found |

## GET /storagebox/{storagebox-id}

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456


    {
      "storagebox":{
        "id":123456,
        "login":"u12345",
        "name":"Backup Server 1",
        "product":"BX60",
        "cancelled":false,
        "locked":false,
        "location":"FSN1",
        "linked_server":123456,
        "paid_until":"2015-10-23",
        "disk_quota":10240000,
        "disk_usage":900,
        "disk_usage_data":500,
        "disk_usage_snapshots":400,
        "webdav":true,
        "samba":true,
        "ssh":true,
        "external_reachability":true,
        "zfs":false,
        "server":"u12345.your-storagebox.de",
        "host_system":"FSN1-BX355"
      }
    }

### Description

Query data of a specific Storage Box

### Request limit

200 requests per 1 hour

### Output

storagebox (Object)

|                                 |                                    |
| ------------------------------- | ---------------------------------- |
| id (Integer)                    | Storage Box ID                     |
| login (String)                  | User name                          |
| name (String)                   | Name of the Storage Box            |
| product (String)                | Product name                       |
| cancelled (Boolean)             | Status of Storage Box cancellation |
| locked (Boolean)                | Status of locking                  |
| location (String)               | Location of Storage Box host       |
| linked_server (Integer)         | Linked server id                   |
| paid_until (String)             | Paid until date                    |
| disk_quota (Integer)            | Total space in MB                  |
| disk_usage (Integer)            | Used space in MB                   |
| disk_usage_data (Integer)       | Used space by data in MB           |
| disk_usage_snapshots (Integer)  | Used space by snapshots in MB      |
| webdav (Boolean)                | Status of WebDAV                   |
| samba (Boolean)                 | Status of Samba                    |
| ssh (Boolean)                   | Status of SSH-Support              |
| external_reachability (Boolean) | Status of external reachability    |
| zfs (Boolean)                   | Status of ZFS directory            |
| server (String)                 | Server                             |
| host_system (String)            | Identifier of Storage Box host     |

### Errors

| Status | Code                 | Description                                   |
| ------ | -------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND | Storage Box with ID {storagebox-id} not found |

## POST /storagebox/{storagebox-id}

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456 -d storagebox_name=backup1


    {
      "storagebox":{
        "id":123456,
        "login":"u12345",
        "name":"Backup Server 1",
        "product":"BX60",
        "cancelled":false,
        "locked":false,
        "location":"FSN1",
        "linked_server":123456,
        "paid_until":"2015-10-23",
        "disk_quota":10240000,
        "disk_usage":900,
        "disk_usage_data":500,
        "disk_usage_snapshots":400,
        "webdav":true,
        "samba":true,
        "ssh":true,
        "external_reachability":true,
        "zfs":false,
        "server":"u12345.your-storagebox.de",
        "host_system":"FSN1-BX355"
      }
    }

### Description

Update a specific Storage Box

### Request limit

1 request per 5 seconds

### Input

| Name                  | Description                     |
| --------------------- | ------------------------------- |
| storagebox_name       | Name of the Storage Box         |
| samba                 | Status of Samba                 |
| webdav                | Status of WebDAV                |
| ssh                   | Status of SSH-Support           |
| external_reachability | Status of external reachability |
| zfs                   | Status of ZFS directory         |

### Output

storagebox (Object)

|                                 |                                    |
| ------------------------------- | ---------------------------------- |
| id (Integer)                    | Storage Box ID                     |
| login (String)                  | User name                          |
| name (String)                   | Name of the Storage Box            |
| product (String)                | Product name                       |
| cancelled (Boolean)             | Status of Storage Box cancellation |
| locked (Boolean)                | Status of locking                  |
| location (String)               | Location of Storage Box host       |
| linked_server (Integer)         | Linked server id                   |
| paid_until (String)             | Paid until date                    |
| disk_quota (Integer)            | Total space in MB                  |
| disk_usage (Integer)            | Used space in MB                   |
| disk_usage_data (Integer)       | Used space by data in MB           |
| disk_usage_snapshots (Integer)  | Used space by snapshots in MB      |
| webdav (Boolean)                | Status of WebDAV                   |
| samba (Boolean)                 | Status of Samba                    |
| ssh (Boolean)                   | Status of SSH-Support              |
| external_reachability (Boolean) | Status of external reachability    |
| zfs (Boolean)                   | Status of ZFS directory            |
| server (String)                 | Server                             |
| host_system (String)            | Identifier of Storage Box host     |

### Errors

| Status | Code                 | Description                                   |
| ------ | -------------------- | --------------------------------------------- |
| 400    | INVALID_INPUT        | Invalid input parameters                      |
| 404    | STORAGEBOX_NOT_FOUND | Storage Box with ID {storagebox-id} not found |

## POST /storagebox/{storagebox-id}/password

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/password -X POST


    {
      "password":"h1cgLgZYJsyGl0JK"
    }

### Description

Reset password of storage box

### Request limit

1 request per 5 seconds

### Output

|                   |                                     |
| ----------------- | ----------------------------------- |
| password (String) | Updated password of the storage box |

### Errors

| Status | Code                 | Description                                   |
| ------ | -------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND | Storage Box with ID {storagebox-id} not found |

## GET /storagebox/{storagebox-id}/snapshot

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/snapshot


    [\
      {\
        "snapshot":{\
          "name":"2015-12-21T12-40-38",\
          "timestamp":"2015-12-21T13:40:38+00:00",\
          "size":400,\
          "filesystem_size":12345,\
          "automatic":false,\
          "comment":"Test-Snapshot"\
        }\
      }\
    ]

### Description

Query snapshots of a specific Storage Box

### Request limit

200 requests per 1 hour

### Output

(Array)snapshot (Object)

|                           |                                                                |
| ------------------------- | -------------------------------------------------------------- |
| name (String)             | Snapshot name                                                  |
| timestamp (String)        | Timestamp of snapshot in UTC                                   |
| size (Integer)            | Snapshot size in MB                                            |
| filesystem_size (Integer) | Size of the Storage Box at creation time of the snapshot in MB |
| automatic (Boolean)       | True if snapshot has been automatically created                |
| comment (String)          | Comment for the snapshot                                       |

### Errors

| Status | Code                 | Description                                   |
| ------ | -------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND | Storage Box with ID {storagebox-id} not found |

## POST /storagebox/{storagebox-id}/snapshot

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/snapshot -X POST


    {
      "snapshot":{
        "name":"2015-12-21T13-13-03",
        "timestamp":"2015-12-21T13:13:03+00:00",
        "size":400
      }
    }

### Description

Create new snapshot of a specific Storage Box

### Request limit

1 request per 5 seconds

### Output

snapshot (Object)

|                    |                              |
| ------------------ | ---------------------------- |
| name (String)      | Snapshot name                |
| timestamp (String) | Timestamp of snapshot in UTC |
| size (Integer)     | Snapshot size in MB          |

### Errors

| Status | Code                    | Description                                   |
| ------ | ----------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND    | Storage Box with ID {storagebox-id} not found |
| 409    | SNAPSHOT_LIMIT_EXCEEDED | Snapshot limit exceeded                       |

## DELETE /storagebox/{storagebox-id}/snapshot/{snapshot-name}

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/snapshot/2015-12-21T13-13-03 \
    -X DELETE

### Description

Delete snapshot

### Request limit

1 request per 5 seconds

### Errors

| Status | Code                 | Description                                   |
| ------ | -------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND | Storage Box with ID {storagebox-id} not found |
| 404    | SNAPSHOT_NOT_FOUND   | Snapshot with name {snapshot-name} not found  |

## POST /storagebox/{storagebox-id}/snapshot/{snapshot-name}

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/snapshot/2015-12-21T13-13-03 \
    -d revert=true

### Description

Revert to snapshot

### Request limit

1 request per 5 seconds

### Input

| Name   | Description                                  |
| ------ | -------------------------------------------- |
| revert | Must be set to "true" to revert the snapshot |

### Errors

| Status | Code                 | Description                                   |
| ------ | -------------------- | --------------------------------------------- |
| 400    | INVALID_INPUT        | Invalid input parameters                      |
| 404    | STORAGEBOX_NOT_FOUND | Storage Box with ID {storagebox-id} not found |
| 404    | SNAPSHOT_NOT_FOUND   | Snapshot with name {snapshot-name} not found  |

## POST /storagebox/{storagebox-id}/snapshot/{snapshot-name}/comment

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/snapshot/2015-12-21T13-13-03/comment -X POST

### Description

Set comment of a specific snapshot

### Request limit

1 request per 5 seconds

### Input

| Name    | Description              |
| ------- | ------------------------ |
| comment | Comment for the snapshot |

### Errors

| Status | Code                 | Description                                   |
| ------ | -------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND | Storage Box with ID {storagebox-id} not found |
| 404    | SNAPSHOT_NOT_FOUND   | Snapshot with name {snapshot-name} not found  |

## GET /storagebox/{storagebox-id}/snapshotplan

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/snapshotplan


    [\
      {\
        "snapshotplan":{\
          "status":"enabled",\
          "minute":5,\
          "hour":12,\
          "day_of_week":2,\
          "day_of_month":null,\
          "month":null,\
          "max_snapshots":2\
        }\
      }\
    ]

### Description

Query data of the snapshot plan of a specific Storage Box

### Request limit

200 requests per 1 hour

### Output

(Array)storagebox (Object)

|                               |                                                                                                            |
| ----------------------------- | ---------------------------------------------------------------------------------------------------------- |
| status (String)               | Status of the snapshot plan                                                                                |
| minute (Integer / null)       | Minute of the execution or null if plan is deactivated                                                     |
| hour (Integer / null)         | Hour of the execution or null if plan is deactivated                                                       |
| day_of_week (Integer / null)  | Weekday of the execution or null if plan is deactivated or value is not set (1 = Monday, ... , 7 = Sunday) |
| day_of_month (Integer / null) | Monthday of the execution or null if plan is deactivated or value is not set (1 = First day of month)      |
| month (Integer / null)        | Month of the execution or null if plan is deactivated or value is not set (1 = January)                    |
| max_snapshots (Integer)       | Maximum number of automatic snapshots of this plan                                                         |

The date and time parameters are in UTC

### Errors

| Status | Code                 | Description                                   |
| ------ | -------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND | Storage Box with ID {storagebox-id} not found |

## POST /storagebox/{storagebox-id}/snapshotplan

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/snapshot/2015-12-21T13-13-03/comment -d status=enabled -d hour=7 -d minute=19


    [\
      {\
        "snapshotplan":{\
          "status":"enabled",\
          "minute":5,\
          "hour":12,\
          "day_of_week":2,\
          "day_of_month":null,\
          "month":null,\
          "max_snapshots":2\
        }\
      }\
    ]

### Description

Edit data of the snapshot plan of a specific Storage Box

### Request limit

1 request per 5 seconds

### Input

| status        | New status of the snapshot plan                           |
| ------------- | --------------------------------------------------------- |
| minute        | Minute of execution. Only required if the plan is enabled |
| hour          | Hour of execution. Only required if the plan is enabled   |
| day_of_week   | Weekday of execution (1 = Monday, ... , 7 = Sunday)       |
| day_of_month  | Monthday of execution (1 = First day of month)            |
| month         | Month of execution (1 = January)                          |
| max_snapshots | Maximum number of automatic snapshots of this plan        |

The date and time parameters are in UTC

### Output

(Array)storagebox (Object)

|                               |                                                                                                            |
| ----------------------------- | ---------------------------------------------------------------------------------------------------------- |
| status (String)               | Status of the snapshot plan                                                                                |
| minute (Integer / null)       | Minute of the execution or null if plan is deactivated                                                     |
| hour (Integer / null)         | Hour of the execution or null if plan is deactivated                                                       |
| day_of_week (Integer / null)  | Weekday of the execution or null if plan is deactivated or value is not set (1 = Monday, ... , 7 = Sunday) |
| day_of_month (Integer / null) | Monthday of the execution or null if plan is deactivated or value is not set (1 = First day of month)      |
| month (Integer / null)        | Month of the execution or null if plan is deactivated or value is not set (1 = January)                    |
| max_snapshots (Integer)       | Maximum number of automatic snapshots of this plan                                                         |

### Errors

| Status | Code                 | Description                                   |
| ------ | -------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND | Storage Box with ID {storagebox-id} not found |

## GET /storagebox/{storagebox-id}/subaccount

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/subaccount


    [\
      {\
        "subaccount":{\
          "username":"u2342-sub1",\
          "accountid":"u2342",\
          "server":"u12345-sub1.your-storagebox.de",\
          "homedirectory":"test",\
          "samba":true,\
          "ssh":true,\
          "external_reachability":true,\
          "webdav":false,\
          "readonly":false,\
          "createtime":"2017-05-24 13:16:45",\
          "comment":"Test-comment"\
        }\
      }\
    ]

### Description

Query data of all sub-accounts of a specific Storage Box

### Request limit

200 requests per 1 hour

### Output

(Array)subaccount (Object)

|                                 |                                       |
| ------------------------------- | ------------------------------------- |
| username (String)               | Username of the sub-account           |
| accountid (String)              | Username of the main user             |
| server (String)                 | Server                                |
| homedirectory (String)          | Homedirectory of the sub-account      |
| samba (Boolean)                 | Status of Samba                       |
| ssh (Boolean)                   | Status of SSH-Support                 |
| external_reachability (Boolean) | Status of external reachability       |
| webdav (Boolean)                | Status of WebDAV                      |
| readonly (Boolean)              | Status of the readonly mode           |
| createtime (String)             | Time when the sub-account was created |
| comment (String)                | Custom comment fot the sub-account    |

### Errors

| Status | Code                 | Description                                   |
| ------ | -------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND | Storage Box with ID {storagebox-id} not found |

## POST /storagebox/{storagebox-id}/subaccount

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/subaccount -d homedirectory=test


    {
      "subaccount":{
        "username":"u2342-sub1",
        "password":"as7udhaisudbasd",
        "accountid":"u2342",
        "server":"u12345-sub1.your-storagebox.de",
        "homedirectory":"test"
      }
    }

### Description

Creates a sub-account

### Request limit

1 request per 5 seconds

### Input

| Name                  | Description                      |
| --------------------- | -------------------------------- |
| homedirectory         | Homedirectory of the sub-account |
| samba                 | Status of Samba                  |
| ssh                   | Status of SSH-Support            |
| external_reachability | Status of external reachability  |
| webdav                | Status of WebDAV                 |
| readonly              | Status of the readonly mode      |
| comment               | Custom comment                   |

### Output

subaccount (Object)

|                        |                                  |
| ---------------------- | -------------------------------- |
| username (String)      | Username of the sub-account      |
| password (String)      | Password of the sub-account      |
| accountid (String)     | Username of the main user        |
| server (String)        | Server                           |
| homedirectory (String) | Homedirectory of the sub-account |

### Errors

| Status | Code                                 | Description                                   |
| ------ | ------------------------------------ | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND                 | Storage Box with ID {storagebox-id} not found |
| 409    | STORAGEBOX_SUBACCOUNT_LIMIT_EXCEEDED | Sub-account limit exceeded                    |

## PUT /storagebox/{storagebox-id}/subaccount/{sub-account-username}

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/subaccount/u2342-sub1 -X PUT -d homedirectory=test

### Description

Update sub-account

### Request limit

1 request per 5 seconds

### Input

| Name                  | Description                      |
| --------------------- | -------------------------------- |
| homedirectory         | Homedirectory of the sub-account |
| samba                 | Status of Samba                  |
| ssh                   | Status of SSH-Support            |
| external_reachability | Status of external reachability  |
| webdav                | Status of WebDAV                 |
| readonly              | Status of the readonly mode      |
| comment               | Custom comment                   |

### Errors

| Status | Code                 | Description                                   |
| ------ | -------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND | Storage Box with ID {storagebox-id} not found |

## DELETE /storagebox/{storagebox-id}/subaccount/{sub-account-username}

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/subaccount/u2342-sub2 \
    -X DELETE

### Description

Delete sub-account

### Request limit

1 request per 5 seconds

### Errors

| Status | Code                            | Description                                   |
| ------ | ------------------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND            | Storage Box with ID {storagebox-id} not found |
| 404    | STORAGEBOX_SUBACCOUNT_NOT_FOUND | Sub-account not found                         |

## POST /storagebox/{storagebox-id}/subaccount/{sub-account-username}/password

    curl -u "user:password" https://robot-ws.your-server.de/storagebox/123456/subaccount/u2342-sub2/password \
    -X POST


    {
      "password":"h1cgLgZYJsyGl0JK"
    }

### Description

Reset password of sub-account

### Request limit

1 request per 5 seconds

### Errors

| Status | Code                            | Description                                   |
| ------ | ------------------------------- | --------------------------------------------- |
| 404    | STORAGEBOX_NOT_FOUND            | Storage Box with ID {storagebox-id} not found |
| 404    | STORAGEBOX_SUBACCOUNT_NOT_FOUND | Sub-account not found                         |

# Firewall

## GET /firewall/{server-id}

    curl -u "user:password" https://robot-ws.your-server.de/firewall/321


    {
      "firewall":{
        "server_ip":"123.123.123.123",
        "server_number":321,
        "status":"active",
        "filter_ipv6":false,
        "whitelist_hos":true,
        "port":"main",
        "rules":{
          "input":{
            "0":{
              "ip_version":"ipv4",
              "name":"rule 1",
              "dst_ip":null,
              "src_ip":"1.1.1.1",
              "dst_port":"80",
              "src_port":null,
              "protocol":null,
              "tcp_flags":null,
              "action":"accept"
            },
            "output":[\
              {\
                "ip_version":null,\
                "name":"Allow all",\
                "dst_ip":null,\
                "src_ip":null,\
                "dst_port":null,\
                "src_port":null,\
                "protocol":null,\
                "tcp_flags":null,\
                "action":"accept"\
              }\
            ]
          }
        }
      }
    }

### Description

Get the firewall configuration for a server

### Request limit

500 requests per 1 hour

### Output

firewall (Object)server_ip (String)Server main IP addressserver_number (Integer)Server IDstatus (String)Status of firewallfilter_ipv6 (Boolean)Flag indicating if IPv6 filter is activewhitelist_hos (Boolean)Flag of Hetzner services whitelistingport (String)Switch port of firewall ('main' or 'kvm')rules (Object)input (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard')           |

output (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard')           |

### Errors

| Status | Code                    | Description                                             |
| ------ | ----------------------- | ------------------------------------------------------- |
| 404    | SERVER_NOT_FOUND        | Server with ID {server-id} not found                    |
| 404    | FIREWALL_PORT_NOT_FOUND | Switch port not found                                   |
| 404    | FIREWALL_NOT_AVAILABLE  | Firewall configuration is not available for this server |

### Deprecations

|                                       |                                                                        |
| ------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated GET /firewall/{server-ip} | The main IPv4 address may be used alternatively to specify the server. |

## POST /firewall/{server-id}

> Before encoding:

    rules[input][0][name]=rule 1 (v4)&
    rules[input][0][ip_version]=ipv4&
    rules[input][0][src_ip]=1.1.1.1&
    rules[input][0][dst_port]=80&
    rules[input][0][action]=accept&
    rules[input][1][name]=Allow MySQL&
    rules[input][1][ip_version]=ipv4&
    rules[input][1][dst_port]=3306&
    rules[input][1][action]=accept&
    rules[output][0][name]=Allow all&
    rules[output][0][ip_version]=ipv4&
    rules[output][0][action]=accept

> After encoding:

    rules%5Binput%5D%5B0%5D%5Bip_version%5D=ipv4&
    rules%5Binput%5D%5B0%5D%5Bname%5D=rule+1+%28v4%29&
    rules%5Binput%5D%5B0%5D%5Bsrc_ip%5D=1.1.1.1&
    rules%5Binput%5D%5B0%5D%5Bdst_port%5D=80&
    rules%5Binput%5D%5B0%5D%5Baction%5D=accept&
    rules%5Binput%5D%5B1%5D%5Bip_version%5D=ipv4&
    rules%5Binput%5D%5B1%5D%5Bname%5D=Allow+MySQL&
    rules%5Binput%5D%5B1%5D%5Bdst_port%5D=3306&
    rules%5Binput%5D%5B1%5D%5Baction%5D=accept
    rules%5Boutput%5D%5B0%5D%5Bname%5D%3DAllow+all%26%0A
    rules%5Boutput%5D%5B0%5D%5Bip_version%5D%3Dipv4%26%0A
    rules%5Boutput%5D%5B0%5D%5Baction%5D%3Daccept


    curl -u "user:password" https://robot-ws.your-server.de/firewall/321 \
    --data-urlencode 'status=active' \
    --data-urlencode 'whitelist_hos=true' \
    --data-urlencode 'rules[input][0][name]=rule 1' \
    --data-urlencode 'rules[input][0][ip_version]=ipv4' \
    --data-urlencode 'rules[input][0][src_ip]=1.1.1.1' \
    --data-urlencode 'rules[input][0][dst_port]=80' \
    --data-urlencode 'rules[input][0][action]=accept' \
    --data-urlencode 'rules[input][1][name]=Allow MySQL' \
    --data-urlencode 'rules[input][1][ip_version]=ipv4' \
    --data-urlencode 'rules[input][1][dst_port]=3306' \
    --data-urlencode 'rules[input][1][action]=accept' \
    --data-urlencode 'rules[output][0][name]=Allow all' \
    --data-urlencode 'rules[output][0][action]=accept'


    {
      "firewall":{
        "server_ip":"123.123.123.123",
        "server_number":321,
        "status":"in process",
        "filter_ipv6":false,
        "whitelist_hos":true,
        "port":"main",
        "rules":{
          "input":[\
            {\
              "ip_version":"ipv4",\
              "name":"rule 1",\
              "dst_ip":null,\
              "src_ip":"1.1.1.1",\
              "dst_port":"80",\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            },\
            {\
              "ip_version":"ipv4",\
              "name":"Allow MySQL",\
              "dst_ip":null,\
              "src_ip":null,\
              "dst_port":"3306",\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            }\
          ],
          "output":[\
            {\
              "ip_version":null,\
              "name":"Allow all",\
              "dst_ip":null,\
              "src_ip":null,\
              "dst_port":null,\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            }\
          ]
        }
      }
    }

### Description

Apply a new firewall configuration

### Request limit

500 requests per 1 hour

### Input

| Name          | Description                                                            |
| ------------- | ---------------------------------------------------------------------- |
| status        | Change the status of the firewall ('active' or 'disabled')             |
| filter_ipv6   | Activate or deactivate the IPv6 filter ('true' or 'false', optional)   |
| whitelist_hos | Change the flag of Hetzner services whitelisting ('true' or 'false')   |
| rules         | Firewall rules                                                         |
| template_id   | Template ID (not possible in combination with whitelist_hos and rules) |

### Rule data

| Name       | Description                                                                  |
| ---------- | ---------------------------------------------------------------------------- |
| name       | Name for rule                                                                |
| ip_version | IP version, ('ipv4', 'ipv6')                                                 |
| dst_ip     | Destination IPv4 address (only usable in combination with ip_version 'ipv4') |
| src_ip     | Source IPv4 address (only usable combination with ip_version 'ipv4')         |
| dst_port   | Destination TCP/UDP port                                                     |
| src_port   | Source TCP/UDP port                                                          |
| protocol   | Protocol ('tcp', 'udp', 'gre', 'icmp', 'ipip', 'ah', 'esp')                  |
| tcp_flags  | TCP flags                                                                    |
| action     | Action ('discard', 'accept')                                                 |

- All rule data except 'action' is optional
- Omitted rule fields will have the value 'null' and will act like a wildcard.
- Parameter 'rules' must be an array with the following structure:
- `rules[{direction}][{rule index}][{rule field}]={rule field data}`
- The direction can be specified as 'input' to filter incoming packets and 'output' to filter outgoing packets.

Please note that all data needs to be encoded as 'application/x-www-form-urlencoded'.

### Limitations IPv6

- It is not possible to filter the ICMPv6 protocol. ICMPv6 traffic to and from the server is always allowed.
- For rules with IP version IPv6 or no specification of IP version, it is not possible to filter on destination and source IP address.
- Without specifying the IP version, it is not possible to filter on a specific protocol.

### Output

firewall (Object)server_ip (String)Server main IP addressserver_number (Integer)Server IDstatus (String)Status of firewallfilter_ipv6 (Boolean)Flag indicating if IPv6 filter is activewhitelist_hos (Boolean)Flag of Hetzner services whitelistingport (String)Switch port of firewall ('main' or 'kvm')rules (Object)input (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard', required) |

output (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard', required) |

### Errors

| Status | Code                         | Description                                                          |
| ------ | ---------------------------- | -------------------------------------------------------------------- |
| 404    | SERVER_NOT_FOUND             | Server with ID {server-id} not found                                 |
| 404    | FIREWALL_PORT_NOT_FOUND      | Switch port not found                                                |
| 404    | FIREWALL_NOT_AVAILABLE       | Firewall configuration is not available for this server              |
| 404    | FIREWALL_TEMPLATE_NOT_FOUND  | Template with ID {template_id} not found                             |
| 409    | FIREWALL_IN_PROCESS          | The firewall cannot be updated because a update is currently running |
| 409    | FIREWALL_RULE_LIMIT_EXCEEDED | The firewall rule limit is exceeded                                  |
| 409    | FIREWALL_CANNOT_BE_DISABLED  | The firewall cannot be disabled because internal rules are set       |

### Deprecations

|                                        |                                                                        |
| -------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated POST /firewall/{server-ip} | The main IPv4 address may be used alternatively to specify the server. |

## DELETE /firewall/{server-id}

    curl -u "user:password" https://robot-ws.your-server.de/firewall/321 -X DELETE


    {
      "firewall":{
        "server_ip":"123.123.123.123",
        "server_number":321,
        "status":"in process",
        "filter_ipv6":false,
        "whitelist_hos":true,
        "port":"main",
        "rules":{

        }
      }
    }

### Description

Clear firewall configuration of a server

### Request limit

500 requests per 1 hour

### Input

No input

### Output

firewall (Object)server_ip (String)Server main IP addressserver_number (Integer)Server IDstatus (String)Status of firewallfilter_ipv6 (Boolean)Flag indicating if IPv6 filter is activewhitelist_hos (Boolean)Flag of Hetzner services whitelistingport (String)Switch port of firewall ('main' or 'kvm')rules (Object)input (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard')           |

output (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard')           |

### Errors

| Status | Code                         | Description                                                          |
| ------ | ---------------------------- | -------------------------------------------------------------------- |
| 404    | SERVER_NOT_FOUND             | Server with ID {server-id} not found                                 |
| 404    | FIREWALL_PORT_NOT_FOUND      | Switch port not found                                                |
| 404    | FIREWALL_NOT_AVAILABLE       | Firewall configuration is not available for this server              |
| 409    | FIREWALL_IN_PROCESS          | The firewall cannot be updated because a update is currently running |
| 409    | FIREWALL_RULE_LIMIT_EXCEEDED | The firewall rule limit is exceeded                                  |
| 409    | FIREWALL_CANNOT_BE_DISABLED  | The firewall cannot be disabled because internal rules are set       |

### Deprecations

|                                          |                                                                        |
| ---------------------------------------- | ---------------------------------------------------------------------- |
| @deprecated DELETE /firewall/{server-ip} | The main IPv4 address may be used alternatively to specify the server. |

## GET /firewall/template

    curl -u "user:password" https://robot-ws.your-server.de/firewall/template


    [\
      {\
        "firewall_template":{\
          "id":1,\
          "name":"My template",\
          "filter_ipv6":false,\
          "whitelist_hos":true,\
          "is_default":true\
        }\
      },\
      {\
        "firewall_template":{\
          "id":2,\
          "name":"My second template",\
          "filter_ipv6":false,\
          "whitelist_hos":true,\
          "is_default":false\
        }\
      }\
    ]

### Description

Get list of available firewall templates

### Request limit

500 requests per 1 hour

### Output

(Array)firewall_template (Object)

|                         |                                                                   |
| ----------------------- | ----------------------------------------------------------------- |
| id (Integer)            | ID of firewall template                                           |
| name (String)           | Name of firewall template                                         |
| filter_ipv6 (Boolean)   | Flag indicating if IPv6 filter is active                          |
| whitelist_hos (Boolean) | Flag of Hetzner services whitelisting                             |
| is_default (Boolean)    | If true the template is selected by default in the Robot webpanel |

### Errors

| Status | Code      | Description                 |
| ------ | --------- | --------------------------- |
| 404    | NOT_FOUND | No firewall templates found |

## POST /firewall/template

    curl -u "user:password" https://robot-ws.your-server.de/firewall/template \
    --data-urlencode 'name=My new template' \
    --data-urlencode 'filter_ipv6=false' \
    --data-urlencode 'whitelist_hos=true' \
    --data-urlencode 'is_default=false' \
    --data-urlencode 'rules[input][0][name]=rule 1' \
    --data-urlencode 'rules[input][0][ip_version]=ipv4' \
    --data-urlencode 'rules[input][0][src_ip]=1.1.1.1' \
    --data-urlencode 'rules[input][0][dst_port]=80' \
    --data-urlencode 'rules[input][0][action]=accept' \
    --data-urlencode 'rules[input][1][name]=Allow MySQL' \
    --data-urlencode 'rules[input][1][ip_version]=ipv4' \
    --data-urlencode 'rules[input][1][dst_port]=3306' \
    --data-urlencode 'rules[input][1][action]=accept'
    --data-urlencode 'rules[output][0][name]=Allow all' \
    --data-urlencode 'rules[output][0][action]=accept'


    {
      "firewall_template":{
        "id":123,
        "filter_ipv6":false,
        "whitelist_hos":true,
        "is_default":false,
        "rules":{
          "input":[\
            {\
              "ip_version":"ipv4",\
              "name":"rule 1",\
              "dst_ip":null,\
              "src_ip":"1.1.1.1",\
              "dst_port":"80",\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            },\
            {\
              "ip_version":"ipv4",\
              "name":"Allow MySQL",\
              "dst_ip":null,\
              "src_ip":null,\
              "dst_port":"3306",\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            }\
          ],
          "output":[\
            {\
              "ip_version":null,\
              "name":"Allow all",\
              "dst_ip":null,\
              "src_ip":null,\
              "dst_port":null,\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            }\
          ]
        }
      }
    }

### Description

Create a new firewall template

### Request limit

500 requests per 1 hour

### Input

| Name          | Description                                                          |
| ------------- | -------------------------------------------------------------------- |
| name          | Template name                                                        |
| filter_ipv6   | Activate or deactivate the IPv6 filter ('true' or 'false', optional) |
| whitelist_hos | Flag of Hetzner services whitelisting                                |
| is_default    | If true the template is selected by default in the Robot webpanel    |
| rules         | Firewall rules                                                       |

Details about the 'rules' parameter are described at [POST /firewall/{server-id}](#post-firewall-server-id)

### Output

firewall_template (Object)id (Integer)ID of firewall templatename (String)Name of firewall templatefilter_ipv6 (Boolean)Flag indicating if IPv6 filter is activewhitelist_hos (Boolean)Flag of Hetzner services whitelistingis_default (Boolean)If true the template is selected by default in the Robot webpanelrules (Object)input (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard')           |

output (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard')           |

## GET /firewall/template/{template-id}

    curl -u "user:password" https://robot-ws.your-server.de/firewall/template/123


    {
      "firewall_template":{
        "id":123,
        "filter_ipv6":false,
        "whitelist_hos":true,
        "is_default":false,
        "rules":{
          "input":[\
            {\
              "ip_version":"ipv4",\
              "name":"rule 1",\
              "dst_ip":null,\
              "src_ip":"1.1.1.1",\
              "dst_port":"80",\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            },\
            {\
              "ip_version":"ipv4",\
              "name":"Allow MySQL",\
              "dst_ip":null,\
              "src_ip":null,\
              "dst_port":"3306",\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            }\
          ],
          "output":[\
            {\
              "ip_version":null,\
              "name":"Allow all",\
              "dst_ip":null,\
              "src_ip":null,\
              "dst_port":null,\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            }\
          ]
        }
      }
    }

### Description

Get a specific firewall template

### Request limit

500 requests per 1 hour

### Output

firewall_template (Object)id (Integer)ID of firewall templatename (String)Name of firewall templatefilter_ipv6 (Boolean)Flag indicating if IPv6 filter is activewhitelist_hos (Boolean)Flag of Hetzner services whitelistingis_default (Boolean)If true the template is selected by default in the Robot webpanelrules (Object)input (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard')           |

output (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard')           |

### Errors

| Status | Code      | Description                 |
| ------ | --------- | --------------------------- |
| 404    | NOT_FOUND | Firewall template not found |

## POST /firewall/template/{template-id}

    curl -u "user:password" https://robot-ws.your-server.de/firewall/template/123 \
    --data-urlencode 'name=My new template' \
    --data-urlencode 'filter_ipv6=false' \
    --data-urlencode 'whitelist_hos=true' \
    --data-urlencode 'is_default=false' \
    --data-urlencode 'rules[input][0][name]=rule 1' \
    --data-urlencode 'rules[input][0][ip_version]=ipv4' \
    --data-urlencode 'rules[input][0][src_ip]=1.1.1.1' \
    --data-urlencode 'rules[input][0][dst_port]=80' \
    --data-urlencode 'rules[input][0][action]=accept' \
    --data-urlencode 'rules[input][1][name]=Allow MySQL' \
    --data-urlencode 'rules[input][1][ip_version]=ipv4' \
    --data-urlencode 'rules[input][1][dst_port]=3306' \
    --data-urlencode 'rules[input][1][action]=accept' \
    --data-urlencode 'rules[input][2][name]=Allow HTTPS' \
    --data-urlencode 'rules[input][2][ip_version]=ipv4' \
    --data-urlencode 'rules[input][2][dst_port]=443' \
    --data-urlencode 'rules[input][2][protocol]=tcp' \
    --data-urlencode 'rules[input][2][action]=accept' \
    --data-urlencode 'rules[output][0][name]=Allow all' \
    --data-urlencode 'rules[output][0][action]=accept'


    {
      "firewall_template":{
        "id":123,
        "filter_ipv6":false,
        "whitelist_hos":true,
        "is_default":false,
        "rules":{
          "input":[\
            {\
              "ip_version":"ipv4",\
              "name":"rule 1",\
              "dst_ip":null,\
              "src_ip":"1.1.1.1",\
              "dst_port":"80",\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            },\
            {\
              "ip_version":"ipv4",\
              "name":"Allow MySQL",\
              "dst_ip":null,\
              "src_ip":null,\
              "dst_port":"3306",\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            },\
            {\
              "ip_version":"ipv4",\
              "name":"Allow HTTPS",\
              "dst_ip":null,\
              "src_ip":null,\
              "dst_port":"443",\
              "src_port":null,\
              "protocol":"tcp",\
              "tcp_flags":null,\
              "action":"accept"\
            }\
          ],
          "output":[\
            {\
              "ip_version":null,\
              "name":"Allow all",\
              "dst_ip":null,\
              "src_ip":null,\
              "dst_port":null,\
              "src_port":null,\
              "protocol":null,\
              "tcp_flags":null,\
              "action":"accept"\
            }\
          ]
        }
      }
    }

### Description

Update a firewall template

### Request limit

500 requests per 1 hour

### Input

| Name          | Description                                                          |
| ------------- | -------------------------------------------------------------------- |
| name          | Template name                                                        |
| filter_ipv6   | Activate or deactivate the IPv6 filter ('true' or 'false', optional) |
| whitelist_hos | Flag of Hetzner services whitelisting                                |
| is_default    | If true the template is selected by default in the Robot webpanel    |
| rules         | Firewall rules                                                       |

Details about the 'rules' parameter are described at [POST /firewall/{server-id}](#post-firewall-server-id)

### Output

firewall_template (Object)id (Integer)ID of firewall templatename (String)Name of firewall templatefilter_ipv6 (Boolean)Flag indicating if IPv6 filter is activewhitelist_hos (Boolean)Flag of Hetzner services whitelistingis_default (Boolean)If true the template is selected by default in the Robot webpanelrules (Object)input (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard')           |

output (Array)(Object)

|                     |                                                          |
| ------------------- | -------------------------------------------------------- |
| ip_version (String) | Internet protocol version ('ipv4' or 'ipv6')             |
| name (String)       | Rule name                                                |
| dst_ip (String)     | Destination IP address or subnet address (CIDR notation) |
| src_ip (String)     | Source IP address or subnet address (CIDR notation)      |
| dst_port (String)   | Destination port or port range                           |
| src_port (String)   | Source port or port range                                |
| protocol (String)   | Protocol above IP layer                                  |
| tcp_flags (String)  | TCP flag or logical combination of flags                 |
| action (String)     | Action if rule matches ('accept' or 'discard')           |

### Errors

| Status | Code      | Description                 |
| ------ | --------- | --------------------------- |
| 404    | NOT_FOUND | Firewall template not found |

## DELETE /firewall/template/{template-id}

    curl -u "user:password" https://robot-ws.your-server.de/firewall/template/123 -X DELETE

### Description

Delete a firewall template

### Request limit

500 requests per 1 hour

### Input

No input

### Output

No output

### Errors

| Status | Code      | Description                 |
| ------ | --------- | --------------------------- |
| 404    | NOT_FOUND | Firewall template not found |

# vSwitch

## GET /vswitch

    curl -u "user:password" https://robot-ws.your-server.de/vswitch


    [\
      {\
        "id":1234,\
        "name":"vswitch 1234",\
        "vlan":4000,\
        "cancelled":false\
      },\
      {\
        "id":4321,\
        "name":"vswitch test",\
        "vlan":4001,\
        "cancelled":false\
      }\
    ]

### Description

Query data of all vSwitches

### Request limit

500 requests per 1 hour

### Output

(Array)(Object)

|                     |                     |
| ------------------- | ------------------- |
| id (Integer)        | vSwitch ID          |
| name (String)       | vSwitch name        |
| vlan (Integer)      | VLAN ID             |
| cancelled (Boolean) | Cancellation status |

## POST /vswitch

    curl -u "user:password" https://robot-ws.your-server.de/vswitch \
    --data-urlencode 'vlan=4000' \
    --data-urlencode 'name=my vSwitch'


    {
      "id":4321,
      "name":"my vSwitch",
      "vlan":4000,
      "cancelled":false,
      "server":[\
    \
      ],
      "subnet":[\
    \
      ],
      "cloud_network":[\
    \
      ]
    }

### Description

Create a new vSwitch

### Request limit

100 requests per 1 hour

### Input

| Name | Description  |
| ---- | ------------ |
| name | vSwitch name |
| vlan | VLAN ID      |

### Output

id (Integer)vSwitch IDname (String)vSwitch namevlan (Integer)VLAN IDcancelled (Boolean)Cancellation statusserver (Array)(Object)

|                          |                                                                     |
| ------------------------ | ------------------------------------------------------------------- |
| server_ip                | Server main IP address                                              |
| server_ipv6_net (String) | Server main IPv6 net address                                        |
| server_number            | Server ID                                                           |
| status                   | Status of vSwitch for this server ("ready", "in process", "failed") |

subnet (Array)(Object)

|                  |                              |
| ---------------- | ---------------------------- |
| ip (String)      | IP address                   |
| mask (Integer)   | Subnet mask in CIDR notation |
| gateway (String) | Gateway                      |

cloud_network (Array)(Object)

|                  |                              |
| ---------------- | ---------------------------- |
| id (Integer)     | Cloud network ID             |
| ip (String)      | IP address                   |
| mask (Integer)   | Subnet mask in CIDR notation |
| gateway (String) | Gateway                      |

### Errors

| Status | Code                  | Description                               |
| ------ | --------------------- | ----------------------------------------- |
| 400    | INVALID_INPUT         | Invalid input parameters                  |
| 409    | VSWITCH_LIMIT_REACHED | The maximum count of vSwitches is reached |

## GET /vswitch/{vswitch-id}

    curl -u "user:password" https://robot-ws.your-server.de/vswitch/4321


    {
      "id":4321,
      "name":"my vSwitch",
      "vlan":4000,
      "cancelled":false,
      "server":[\
        {\
          "server_ip":"123.123.123.123",\
          "server_ipv6_net":"2a01:4f8:111:4221::",\
          "server_number":321,\
          "status":"ready"\
        },\
        {\
          "server_ip":"123.123.123.124",\
          "server_ipv6_net":"2a01:4f8:111:4221::",\
          "server_number":421,\
          "status":"ready"\
        }\
      ],
      "subnet":[\
        {\
          "ip":"213.239.252.48",\
          "mask":29,\
          "gateway":"213.239.252.49"\
        }\
      ],
      "cloud_network":[\
        {\
          "id":123,\
          "ip":"10.0.2.0",\
          "mask":24,\
          "gateway":"10.0.2.1"\
        }\
      ]
    }

### Description

Query data of a specific vSwitch

### Request limit

500 requests per 1 hour

### Output

id (Integer)vSwitch IDname (String)vSwitch namevlan (Integer)VLAN IDcancelled (Boolean)Cancellation statusserver (Array)(Object)

|                          |                                                                     |
| ------------------------ | ------------------------------------------------------------------- |
| server_ip                | Server main IP address                                              |
| server_ipv6_net (String) | Server main IPv6 net address                                        |
| server_number            | Server ID                                                           |
| status                   | Status of vSwitch for this server ("ready", "in process", "failed") |

subnet (Array)(Object)

|                  |                              |
| ---------------- | ---------------------------- |
| ip (String)      | IP address                   |
| mask (Integer)   | Subnet mask in CIDR notation |
| gateway (String) | Gateway                      |

cloud_network (Array)(Object)

|                  |                              |
| ---------------- | ---------------------------- |
| id (Integer)     | Cloud network ID             |
| ip (String)      | IP address                   |
| mask (Integer)   | Subnet mask in CIDR notation |
| gateway (String) | Gateway                      |

### Errors

| Status | Code      | Description       |
| ------ | --------- | ----------------- |
| 404    | NOT_FOUND | vSwitch not found |

## POST /vswitch/{vswitch-id}

    curl -u "user:password" https://robot-ws.your-server.de/vswitch/4321 \
    --data-urlencode 'name=my new name'
    --data-urlencode 'vlan=4001'

### Description

Change the name or the VLAN ID of a vSwitch

### Request limit

100 requests per 1 hour

### Input

| Name | Description  |
| ---- | ------------ |
| name | vSwitch name |
| vlan | VLAN ID      |

### Output

No output

### Errors

| Status | Code                    | Description                                                         |
| ------ | ----------------------- | ------------------------------------------------------------------- |
| 400    | INVALID_INPUT           | Invalid input parameters                                            |
| 404    | NOT_FOUND               | vSwitch not found                                                   |
| 409    | VSWITCH_IN_PROCESS      | The vSwitch cannot be updated because a update is currently running |
| 409    | VSWITCH_VLAN_NOT_UNIQUE | The vSwitch cannot be updated because of a conflicting VLAN ID      |

## DELETE /vswitch/{vswitch-id}

    curl -u "user:password" https://robot-ws.your-server.de/vswitch/4321 -X DELETE \
    --data-urlencode 'cancellation_date=2018-06-30'

### Description

Cancel a vSwitch

### Request limit

100 requests per 1 hour

### Input

| Name              | Description                                                                                      |
| ----------------- | ------------------------------------------------------------------------------------------------ |
| cancellation_date | Date to which the vSwitch should be cancelled (format yyyy-MM-dd) or "now" to cancel immediately |

### Output

No output

### Errors

| Status | Code          | Description                      |
| ------ | ------------- | -------------------------------- |
| 400    | INVALID_INPUT | Invalid input parameters         |
| 404    | NOT_FOUND     | vSwitch not found                |
| 409    | CONFLICT      | The vSwitch is already cancelled |

## POST /vswitch/{vswitch-id}/server

    curl -u "user:password" https://robot-ws.your-server.de/vswitch/4321/server \
    --data-urlencode 'server[]=123.123.123.123'
    --data-urlencode 'server[]=123.123.123.124'

### Description

Add one more servers to a vSwitch

### Request limit

100 requests per 1 hour

### Input

| Name   | Description                                                                       |
| ------ | --------------------------------------------------------------------------------- |
| server | One server identifier or array of server identifiers (server_number or server_ip) |

### Output

No output

### Errors

| Status | Code                             | Description                                                         |
| ------ | -------------------------------- | ------------------------------------------------------------------- |
| 400    | INVALID_INPUT                    | Invalid input parameters                                            |
| 404    | NOT_FOUND                        | vSwitch not found                                                   |
| 404    | SERVER_NOT_FOUND                 | A submitted server is not found                                     |
| 404    | VSWITCH_NOT_AVAILABLE            | The vSwitch feature is not available for a submitted server         |
| 409    | VSWITCH_IN_PROCESS               | The vSwitch cannot be updated because a update is currently running |
| 409    | VSWITCH_VLAN_NOT_UNIQUE          | The vSwitch cannot be updated because of a conflicting VLAN ID      |
| 409    | VSWITCH_SERVER_LIMIT_REACHED     | The maximum number of servers is reached for this vSwitch           |
| 409    | VSWITCH_PER_SERVER_LIMIT_REACHED | The maximum number of vSwitches is reached for a submitted server   |

## DELETE /vswitch/{vswitch-id}/server

    curl -u "user:password" https://robot-ws.your-server.de/vswitch/4321/server \
    --data-urlencode 'server[]=123.123.123.123'
    --data-urlencode 'server[]=123.123.123.124'

### Description

Delete one more servers from a vSwitch

### Request limit

100 requests per 1 hour

### Input

| Name   | Description                                                                       |
| ------ | --------------------------------------------------------------------------------- |
| server | One server identifier or array of server identifiers (server_number or server_ip) |

### Output

No output

### Errors

| Status | Code               | Description                                                         |
| ------ | ------------------ | ------------------------------------------------------------------- |
| 400    | INVALID_INPUT      | Invalid input parameters                                            |
| 404    | NOT_FOUND          | vSwitch not found                                                   |
| 404    | SERVER_NOT_FOUND   | A submitted server is not found                                     |
| 409    | VSWITCH_IN_PROCESS | The vSwitch cannot be updated because a update is currently running |

# PHP Client

There is a PHP client library available for download at [https://robot.hetzner.com/downloads/robot-client.zip](https://robot.hetzner.com/downloads/robot-client.zip)
. You need PHP with [libcurl](http://www.php.net/curl)
support to use it. \`\`\`php <?php require 'RobotRestClient.class.php'; require 'RobotClientException.class.php'; require 'RobotClient.class.php';

$robot = new RobotClient('https://robot-ws.your-server.de', 'login', 'password');

// retrieve all failover ips $results = $robot->failoverGet();

foreach ($results as $result) { echo $result->failover->ip . "\\n"; echo $result->failover->server_ip . "\\n"; echo $result->failover->active_server_ip . "\\n"; }

// retrieve a specific failover ip $result = $robot->failoverGet('123.123.123.123');

echo $result->failover->ip . "\\n"; echo $result->failover->server_ip . "\\n"; echo $result->failover->active_server_ip . "\\n";

// switch routing try { $robot->failoverRoute('123.123.123.123', '213.133.104.190'); } catch (RobotClientException $e) { echo $e->getMessage() . "\\n"; } \`\`\`
