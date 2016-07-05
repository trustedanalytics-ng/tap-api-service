# TAP Api Service
Api service is a microservice developed to be part of TAP platform.
It is used as a main gateway to whole TAP platform.

It allows you to manage:
* offerings
* applications
* services
* users

## REQUIREMENTS

### Dependencies
This component depends on and communicates with:
* [tap-catalog](https://github.com/intel-data/tap-catalog)
* [tap-template-repository](https://github.com/intel-data/tap-template-repository)
* [tap-blob-store](https://github.com/intel-data/tap-blob-store)
* [tap-container-broker](https://github.com/intel-data/tap-container-broker)
* [tap-image-factory](https://github.com/intel-data/tap-image-factory)
* [user-management](https://github.com/intel-data/user-management)

### Compilation
* git (for pulling repository)
* go >= 1.6

## Compilation
To build and run project:
```bash
  git clone https://github.com/intel-data/tap-api-service
  cd tap-api-service
  make build_anywhere
```
Binaries will be available in ./application directory.

## USAGE
Api Service endpoints are documented in swagger.yaml file.
Below you can find sample usages of Api Service.

### Getting OAuth2 token
To get OAuth2 token, you need to call login endpoint with basic auth:
```bash
curl http://$API_SERVICE_IP/api/v1/login -u admin:password
```
response:
```json
{
   "access_token":"eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiI4ZTQwZDI4Ni0wNTMyLTQ5M2ItYWNiMC02NjQ1ZGFlN2I2YmEiLCJzdWIiOiIzOGQ1ZDY4OC0yYWUzLTQ5ZjItYWVlOC0xMTM4YTgxMjVlNGEiLCJzY29wZSI6WyJvcGVuaWQiLCJ1YWEuYWRtaW4iLCJwYXNzd29yZC53cml0ZSIsInRhcC51c2VyIiwidGFwLmFkbWluIl0sImNsaWVudF9pZCI6ImNvbnNvbGVTdmMiLCJjaWQiOiJjb25zb2xlU3ZjIiwiYXpwIjoiY29uc29sZVN2YyIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiIzOGQ1ZDY4OC0yYWUzLTQ5ZjItYWVlOC0xMTM4YTgxMjVlNGEiLCJvcmlnaW4iOiJ1YWEiLCJ1c2VyX25hbWUiOiJhZG1pbiIsImVtYWlsIjoiYWRtaW5AZXhhbXBsZS5jb20iLCJhdXRoX3RpbWUiOjE0NzczOTEwNzEsInJldl9zaWciOiJjMTJiNTk4YiIsImlhdCI6MTQ3NzM5MTA3MiwiZXhwIjoxNDc3NDM0MjcyLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwODAvdWFhL29hdXRoL3Rva2VuIiwiemlkIjoidWFhIiwiYXVkIjpbImNvbnNvbGVTdmMiLCJvcGVuaWQiLCJ1YWEiLCJwYXNzd29yZCIsInRhcCJdfQ.oAgDNuzYudgB7LGyjBF2ccJpTt51Ah4yhE38-iFXRkPMjVmmHhzeCLI-hlWd_SrjS7ERHaCU9GoFrJOrRdbX9kycDNwazWX1X85-SwXscM5ZiHqE9KVwPunX3tEc7A5Qf7ojTJt2Uph1U8-UZTFRFra1-7pdvBgZ0Fk9NHMF_vXDXJAkLyZSwv6ijHe1sJWf_07szFJxMn6Z8DgHLQ-oP3rmy_b0ukb3O8_WpZf7Z391DRUzNMGKoWYN25nD6nDQHcTN1lx9TbZuGTC0HgcNSmYiDm2qU7rZ5cQndqM7J3fVGtNXZCLHBbRVX6KzbbQMkO1dhnM4sx6MqQimizS0Og",
   "refresh_token":"eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiI3MDBiNWViNi1iYjMyLTQwODgtYmNmNS0yZTZkMTc2NmEyNjMtciIsInN1YiI6IjM4ZDVkNjg4LTJhZTMtNDlmMi1hZWU4LTExMzhhODEyNWU0YSIsInNjb3BlIjpbIm9wZW5pZCIsInVhYS5hZG1pbiIsInBhc3N3b3JkLndyaXRlIiwidGFwLnVzZXIiLCJ0YXAuYWRtaW4iXSwiaWF0IjoxNDc3MzkxMDcyLCJleHAiOjE0Nzk5ODMwNzIsImNpZCI6ImNvbnNvbGVTdmMiLCJjbGllbnRfaWQiOiJjb25zb2xlU3ZjIiwiaXNzIjoiaHR0cDovL2xvY2FsaG9zdDo4MDgwL3VhYS9vYXV0aC90b2tlbiIsInppZCI6InVhYSIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfbmFtZSI6ImFkbWluIiwib3JpZ2luIjoidWFhIiwidXNlcl9pZCI6IjM4ZDVkNjg4LTJhZTMtNDlmMi1hZWU4LTExMzhhODEyNWU0YSIsInJldl9zaWciOiJjMTJiNTk4YiIsImF1ZCI6WyJjb25zb2xlU3ZjIiwib3BlbmlkIiwidWFhIiwicGFzc3dvcmQiLCJ0YXAiXX0.K7fUWwPnU0iFIT-FS8mpGZrOIov9dFT2QbKr3KCI6hMvb_MCRGVtomZ2uU4Hj5GUCoAnIUg5lcJMmz-eKF9HhDOvvidcf3Ts5HT40CI8ZLL4EnYshZsuc5kxD7AW7kaMJpxhtHpWF6OpEwPOv4QTXOaQLL53zxAHV4vvDTanDS8AfO8BXXy-diRH5kBxbbx3jbKCjlH0-jMryZpB1hASmUYL7NTrnWWYIyfDqnT7U0grac7ZG_r4X-F2Tqzn-_g1WzgcLPZ5o8yKrhLYswSyNH5RYr_B6CgHN8LdKKEEniLaNx18Kbjfy_FaXZM4gQFmgI8luwBfh2usRTNIzWa8iQ",
   "token_type":"bearer",
   "expires_in":43199,
   "scope":"openid uaa.admin password.write tap.user tap.admin",
   "jti":"8e40d286-0532-493b-acb0-6645dae7b6ba"
}
```

In following examples we assume that environment variable OAUTH_TOKEN contains access_token:
```bash
OAUTH_TOKEN=`curl http://$API_SERVICE_IP/api/v1/login -u admin:password | jq .access_token`
```

### Offerings
#### Creating offering
Having file [co_nats.json](https://github.com/intel-data/tap-cli/blob/develop/examples/co_nats.json) containing offering definition and `OAUTH_TOKEN` variable from previous request response, you can create offering:
```bash
curl http://$API_SERVICE_IP/api/v1/offerings -X POST -d "@co_nats.json" -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
[
   {
      "id":"4655bc0a-68e8-47cc-6e80-6833e62aac89",
      "name":"nats",
      "description":"NATS is a lightweight cloud messaging system",
      "bindable":true,
      "templateId":"332fe85f-d700-4cc8-43c5-fe9a85e0adc6",
      "state":"READY",
      "plans":[
         {
            "id":"f2164c7b-f86f-4651-59bc-dc7c825a6bd9",
            "name":"free",
            "description":"free",
            "cost":"free",
            "dependencies":null,
            "auditTrail":{
               "createdOn":1477400733,
               "createdBy":"admin",
               "lastUpdatedOn":1477400733,
               "lastUpdateBy":"admin"
            }
         }
      ],
      "auditTrail":{
         "createdOn":1477400733,
         "createdBy":"admin",
         "lastUpdatedOn":1477400733,
         "lastUpdateBy":"admin"
      },
      "metadata":null
   }
]
```

#### Create offering from binary jar archive

Assuming you have binary jar file (binary.jar) you can create new offering from that file. In order to do that one has to:
* Create manifest file [binary_manifest.json](https://github.com/intel-data/tap-api-service/blob/develop/examples/binary_manifest.json)
* Create service description [binary_offering.json](https://github.com/intel-data/tap-api-service/blob/develop/examples/binary_offering.json)
* Use `OAUTH_TOKEN` variable from login request response


```bash
curl http://$API_SERVICE_IP/api/v1/offerings/binary -X POST -F manifest=@examples/binary_manifest.json -F blob=@binary.jar -F offering=@examples/binary_offering.json  -H "Authorization: Bearer $OAUTH_TOKEN"
```

response:
```json
{
    "id":"9de66bb4-43cd-423a-689c-a86f227a0b55",
    "name":"binary-test-name",
    "description":"offering from JAR",
    "bindable":true,
    "templateId":"",
    "state":"DEPLOYING",
    "plans":[
        {
            "id":"622fee4e-189d-415e-6651-71d421a3dfbb",
            "name":"free",
            "description":"free",
            "cost":"free",
            "dependencies":null,
            "auditTrail":{
                "createdOn":1478175812,
                "createdBy":"",
                "lastUpdatedOn":1478175812,
                "lastUpdateBy":""
             }
         }],
    "auditTrail":{
        "createdOn":1478175812,
        "createdBy":"admin",
        "lastUpdatedOn":1478175812,
        "lastUpdateBy":"admin"},
        "metadata":null
}
```

#### Create offering from application

Assuming you have application in Running state you can create new offering from that application. In order to do that one has to:

```bash
curl http://$API_SERVICE_IP/api/v3/offerings/application -X POST -d '{"applicationId":"exampleAppId", "offeringName": "exampleofferingname", "offeringDisplayName": "Example Offering Name", "Description": "Example Description"}' -H "Authorization: Bearer $OAUTH_TOKEN"
```

response:
```json
{
    "id":"9de66bb4-43cd-423a-689c-a86f227a0b55",
    "name":"exampleofferingname",
    "description":"Example Description",
    "bindable":true,
    "templateId":"",
    "state":"READY",
    "plans":[
        {
            "id":"622fee4e-189d-415e-6651-71d421a3dfbb",
            "name":"standard",
            "description":"free",
            "cost":"free",
            "dependencies":null,
            "auditTrail":{
                "createdOn":1478175812,
                "createdBy":"",
                "lastUpdatedOn":1478175812,
                "lastUpdateBy":""
             }
         }],
    "auditTrail":{
        "createdOn":1478175812,
        "createdBy":"admin",
        "lastUpdatedOn":1478175812,
        "lastUpdateBy":"admin"},
        "metadata":null
}
```

#### Listing offerings
```bash
curl http://$API_SERVICE_IP/api/v1/offerings -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
[
    {
        "state": "READY",
        "tags": null,
        "id": "5cc1eb\n26-199c-4e2a-5adb-300e998bfbcf",
        "name": "jupyter",
        "displayName": "Jupyter",
        "provider": "",
        "url": "",
        "description": "Jupyter notebook server with the full SciPy stack + more",
        "version": "",
        "bindable": true
        "metadata": [
            {
                "value": "Jupyter",
                "key": "displayName"
            },
            {
                "value": "http://ipython.org/documentation.html",
                "key": "documentationUrl"
            },
            {
                "value": "",
                "key": "imageUrl"
            },
            {
                "value": "Docker container for the [SciPy stack](../scipystack) and configured Jupyter notebook server.",
                "key": "longDescription"
            },
            {
                "value": "",
                "key": "providerDisplayName"
            },
            {
                "value\n": "https://github.com/ipython/ipython/wiki/Frequently-asked-questions",
                "key": "supportUrl"
            }
        ],
        "offeringPlans": [
            {
                "active": true,
                "id": "fd157cae-0938-\n4a5b-7a13-4f4fcfd013f0",
                "offeringId": "5cc1eb26-199c-4e2a-5adb-300e998bfbcf",
                "description": "free",
                "free": true,
                "name": "free"
            }
        ],
    }
]
```

#### Deleting offering
```bash
curl http://$API_SERVICE_IP/api/v1/offerings/4655bc0a-68e8-47cc-6e80-6833e62aac89 -X DELETE -H "Authorization: Bearer $OAUTH_TOKEN"
```

### Services
When offering is available, a service can be created.

#### Creating service
To create service, you need to provide among others id of previously created offering as `classId` and plan id generated during offering creation in `Metadata` table with key `PLAN_ID`:
```bash
curl http://$API_SERVICE_IP/api/v1/services -X POST -d '{"name":"mynats", "type":"SERVICE", "classId":"4655bc0a-68e8-47cc-6e80-6833e62aac89", "Metadata": [{"Key":"PLAN_ID", "Value":"f2164c7b-f86f-4651-59bc-dc7c825a6bd9"}]}' -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
{
   "id":"a696d5f3-0dd3-4377-6896-3482d512593f",
   "name":"mynats",
   "type":"SERVICE",
   "classId":"4655bc0a-68e8-47cc-6e80-6833e62aac89",
   "bindings":null,
   "metadata":[
      {
         "key":"PLAN_ID",
         "value":"f2164c7b-f86f-4651-59bc-dc7c825a6bd9"
      }
   ],
   "state":"REQUESTED",
   "auditTrail":{
      "createdOn":1477400881,
      "createdBy":"admin",
      "lastUpdatedOn":1477400881,
      "lastUpdateBy":"admin"
   }
}
```

#### Listing services
```bash
curl http://$API_SERVICE_IP/api/v1/services -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
[
   {
      "id":"a696d5f3-0dd3-4377-6896-3482d512593f",
      "name":"mynats",
      "type":"SERVICE",
      "classId":"4655bc0a-68e8-47cc-6e80-6833e62aac89",
      "bindings":null,
      "metadata":[
         {
            "key":"PLAN_ID",
            "value":"f2164c7b-f86f-4651-59bc-dc7c825a6bd9"
         }
      ],
      "state":"RUNNING",
      "auditTrail":{
         "createdOn":1477400881,
         "createdBy":"admin",
         "lastUpdatedOn":1477400894,
         "lastUpdateBy":"admin"
      },
      "serviceName":"nats",
      "planName":"free"
   }
]
```

#### Obtaining service logs
Having service with "RUNNING" state you can retrieve its logs:
```bash
curl http://$API_SERVICE_IP/api/v1/services/a696d5f3-0dd3-4377-6896-3482d512593f/logs -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
{
   "xa696d5f30dd34-3288389905-3fzqy-k-nats":"[1] 2016/10/25 13:08:05.264561 [INF] Starting nats-server version 0.8.1\n[1] 2016/10/25 13:08:05.264588 [INF] Starting http monitor on :8222\n[1] 2016/10/25 13:08:05.264683 [INF] Listening for route connections on 0.0.0.0:6222\n[1] 2016/10/25 13:08:05.264719 [INF] Listening for client connections on 0.0.0.0:4222\n[1] 2016/10/25 13:08:05.264736 [INF] Server is ready\n"
}
```

#### Deleting service
```bash
curl http://$API_SERVICE_IP/api/v1/services/a696d5f3-0dd3-4377-6896-3482d512593f -X DELETE -H "Authorization: Bearer $OAUTH_TOKEN"
```

#### Exposing service
Every native service can be exposed to get external access to it:
```bash
curl http://$API_SERVICE_IP/api/v1/services/a696d5f3-0dd3-4377-6896-3482d512593/expose -X PUT -d '{"exposed": true}' -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN"
```

response:
```json
[
    "orient-2480.daily.gotapaas.eu",
    "orient-2424.daily.gotapaas.eu"
]
```
Exposed addresses will be added to Instance metadata on key 'urls'.


### Applications
There is possibility to push application written in Java, Python, Go or Node.js.

#### Creating application
To create application, you have to provide manifest describing application and .tar.gz package with binaries or source code and run.sh file, which should be starting application:
```bash
curl http://$API_SERVICE_IP/api/v1/applications -F blob=@tapng-sample-java-app.tar.gz -F manifest=@manifest.json -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
{
   "id":"867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6",
   "name":"sample",
   "description":"",
   "imageId":"app_867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6",
   "replication":1,
   "templateId":"52b31177-fef1-49af-7aa4-90382f7be43e",
   "auditTrail":{
      "createdOn":1477556258,
      "createdBy":"admin",
      "lastUpdatedOn":1477556258,
      "lastUpdateBy":"admin"
   },
   "instanceDependencies":null
}
```

Manifest.json used in this example:
```json
{
  "type": "JAVA",
  "name": "sample",
  "instances": 1,
  "bindings": []
}
```

Manifest representation:

Field | Description | Optional
--- | --- | ---
name | application name | false
type | application type [JAVA, GO, NODEJS, PYTHON2.7, PYTHON3.4] | false
instances | number of start application instances (max value: 5) | false
bindings | list of dependent instance IDs which environment variables will be provided to new application | true

#### Listing applications
```bash
curl http://$API_SERVICE_IP/api/v1/applications -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
[
   {
      "id":"867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6",
      "name":"sample",
      "type":"APPLICATION",
      "classId":"867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6",
      "bindings":null,
      "metadata":[
         {
            "key":"APPLICATION_IMAGE_ADDRESS",
            "value":"$repository_uri/app_867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6"
         },
         {
            "key":"urls",
            "value":"sample.84-42.taplab.sclab.intel.com"
         }
      ],
      "state":"RUNNING",
      "auditTrail":{
         "createdOn":1477556267,
         "createdBy":"admin",
         "lastUpdatedOn":1477556282,
         "lastUpdateBy":"admin"
      },
      "replication":1,
      "imageState":"READY",
      "urls":[
         "sample.84-42.taplab.sclab.intel.com"
      ],
      "imageType":"JAVA",
      "memory":"256MB",
      "disk_quota":"1024MB",
      "running_instances":1
   }
]
```
As you can observe, `metadata` field was added to application. It contains information about application like url address or address of it's docker image.

#### Adding binding
Binding instances allows one instance to have credentials of another instance. For example when you bind MySQL service to some application, you allow application to connect to MySQL database with provided credentials.

Having service with id c31e0954-089e-4ed0-4dc2-8f3a41bbd9e2, you can bind it to application with:
```bash
curl http://$API_SERVICE_IP/api/v1/applications/867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6/bindings -X POST -d '{"service_id":"c31e0954-089e-4ed0-4dc2-8f3a41bbd9e2"}' -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN"
```
Listing application, you may observe that application has now field `bindings` with data from bound service:
```bash
curl http://$API_SERVICE_IP/api/v1/applications/867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6 -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
{
   "id":"867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6",
   "name":"sample",
   "type":"APPLICATION",
   "classId":"867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6",
   "bindings":[
      {
         "id":"c31e0954-089e-4ed0-4dc2-8f3a41bbd9e2",
         "data":{
            "MYNATS_MANAGED_BY":"TAP",
            "MYNATS_NATS_PASSWORD":"",
            "MYNATS_NATS_USERNAME":""
         }
      }
   ],
   "metadata":[
      {
         "key":"APPLICATION_IMAGE_ADDRESS",
         "value":"127.0.0.1:30000/app_867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6"
      },
      {
         "key":"urls",
         "value":"sample.84-42.taplab.sclab.intel.com"
      }
   ],
   "state":"RUNNING",
   "auditTrail":{
      "createdOn":1477556267,
      "createdBy":"admin",
      "lastUpdatedOn":1477556750,
      "lastUpdateBy":"admin"
   },
   "replication":1,
   "imageState":"READY",
   "urls":[
      "sample.84-42.taplab.sclab.intel.com"
   ],
   "imageType":"JAVA",
   "memory":"256MB",
   "disk_quota":"1024MB",
   "running_instances":1
}
```

In this case service is bound to application. If you want to bind application to application, you have to provide field `application_id` instead of `service_id` in body.
The same binding functionality for services allowing binding services and applications to services is handled on api/v1/services path.


#### Listing bindings
```bash
curl http://$API_SERVICE_IP/api/v1/applications/867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6/bindings -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
{
   "resources":[
      {
         "entity":{
            "app_guid":"5895a78b-8791-45dd-681f-1888a131a279",
            "service_instance_guid":"c31e0954-089e-4ed0-4dc2-8f3a41bbd9e2",
            "service_instance_name":"mynats"
         }
      }
   ]
}
```

#### Obtaining application logs
```bash
curl http://$API_SERVICE_IP/api/v1/applications/867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6/logs -H "Authorization: Bearer $OAUTH_TOKEN"
```

#### Scaling application
Application can be scaled to the provided kubernetes pod replicas:
```bash
curl http://$API_SERVICE_IP/api/v1/applications/867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6/scale -X PUT -d '{"replicas":3}' -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN"
```

#### Restarting application
```bash
curl http://$API_SERVICE_IP/api/v1/applications/867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6/restart -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN"
```

#### Starting application
```bash
curl http://$API_SERVICE_IP/api/v1/applications/867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6/start -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN"
```

#### Stopping application
```bash
curl http://$API_SERVICE_IP/api/v1/applications/867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6/stop -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN"
```

#### Deleting application
```bash
curl http://$API_SERVICE_IP/api/v1/applications/867bb0c5-f7f8-4dbc-6bb0-0ae19b7238f6 -X DELETE -H "Authorization: Bearer $OAUTH_TOKEN"
```

### Users
Api Service allows management of platform users.

#### Inviting a user
New user can be invited by sending invitation mail with activation link to the address provided. User will be eventually registered when password is set during activation process.
```bash
curl http://$API_SERVICE_IP/api/v1/users/invitations -X POST -d '{"email":"test.user@somedomain.com"}' -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
{
   "state":"NEW",
   "details":"http://console.daily-nokrb-aws.gotapaas.eu/new-account?code=b66d2549-ea68-408a-b32e-bf2940714276"
}
```

#### Listing invitations
```bash
curl http://$API_SERVICE_IP/api/v1/users/invitations -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
[
   "test.user@somedomain.com",
   "intel.data.tests+aether_20161026_102713_609795@gmail.com",
   "intel.data.tests+ubuntuit_dp2_07_20161026_004140_078933@gmail.com"
]
```

#### Deleting user invitation
```bash
curl http://$API_SERVICE_IP/api/v1/users/invitations -X DELETE -d '{"email":"test.user@somedomain.com"}' -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN"
```

#### Listing users
```bash
curl http://$API_SERVICE_IP/api/v1/users -H "Authorization: Bearer $OAUTH_TOKEN"
```
response:
```json
[
   {
      "guid":"4d2e3484-78a7-40b5-a45b-017cf551f892",
      "username":"taptester"
   },
   {
      "guid":"b5816d77-7641-4700-b090-4a38007d307e",
      "username":"admin"
   },
   {
      "guid":"d0b79c45-ebe2-4754-9385-58367b0073c7",
      "username":"intel.data.tests+ubuntuit_dp2_02_20161026_001212_966594@gmail.com"
   }
]
```

#### Changing user password
Password for logged user can be changed providing old and new password:
```bash
curl http://$API_SERVICE_IP/api/v1/users/current/password -X PUT -d '{"current_password":"old123", "new_password":"new123"}' -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN"
```

Changing password revokes OAuth2 token and you need to obtain it again.

#### Deleting user
```bash
curl http://$API_SERVICE_IP/api/v1/users -X DELETE -d '{"email":"test.user@somedomain.com"}' -H "Content-Type: application/json" -H "Authorization: Bearer $OAUTH_TOKEN" -v
```

### CLI resources

#### Linux 32-bit
```bash
curl -O -J -L -H "Authorization: Bearer $OAUTH_TOKEN" http://$API_SERVICE_IP/api/v1/resources/cli/linux32
```

#### Linux 64-bit
```bash
curl -O -J -L -H "Authorization: Bearer $OAUTH_TOKEN" http://$API_SERVICE_IP/api/v1/resources/cli/linux64
```

#### OS X 64-bit
```bash
curl -O -J -L -H "Authorization: Bearer $OAUTH_TOKEN" http://$API_SERVICE_IP/api/v1/resources/cli/macosx64
```

#### Windows 32-bit
```bash
curl -O -J -L -H "Authorization: Bearer $OAUTH_TOKEN" http://$API_SERVICE_IP/api/v1/resources/cli/windows32
```

## Configuration
Following environment variables configure Api Service:

| Variable | Description |
| --- | --- |
| BIND_ADDRESS | address to listen on  |
| PORT | port to listen on |
| DOMAIN | platform domain |
| TAP_VERSION | tap version |
| CLI_VERSION | cli version |
| CORE_ORGANIZATION | core organization name |
| CDH_VERSION | CDH version |
| K8S_VERSION | Kubernetes version |
| BROKER_LOG_LEVEL | logger level [CRITICAL,ERROR,WARNING,NOTICE,INFO,DEBUG] |
| TEMPLATE_REPOSITORY_KUBERNETES_SERVICE_NAME | kubernetes service name of template repository component  |
| TEMPLATE_REPOSITORY_USER | username for template repository |
| TEMPLATE_REPOSITORY_PASS | password for template repository |
| GENERIC_APPLICATION_TEMPLATE_ID | id of default application template |
| CATALOG_KUBERNETES_SERVICE_NAME | kubernetes service name of catalog component  |
| CATALOG_USER | username for catalog |
| CATALOG_PASS | password for catalog |
| CONTAINER_BROKER_KUBERNETES_SERVICE_NAME | kubernetes service name of container broker component  |
| CONTAINER_BROKER_USER | username for container broker |
| CONTAINER_BROKER_PASS | password for container broker |
| BLOB_STORE_KUBERNETES_SERVICE_NAME | kubernetes service name of blob store component  |
| BLOB_STORE_USER | username for blob store |
| BLOB_STORE_PASS | password for blob store |
| IMAGE_FACTORY_KUBERNETES_SERVICE_NAME | kubernetes service name of image factory component |
| IMAGE_FACTORY_USER | username for image factory |
| IMAGE_FACTORY_PASS | password for image factory |
| SSO_TOKEN_URI | user management URI for generating ouath tokens  |
| SSO_CHECK_TOKEN_URI | user management URI for checking oauth tokens |
| SSO_CLIENT | user management oauth client |
| SSO_SECRET | user management oauth secret |
| USER_MANAGEMENT_KUBERNETES_SERVICE_NAME | kubernetes service name of user management component |
| USER_MANAGEMENT_SSL_CERT_FILE_LOCATION | user management certification file location |
| USER_MANAGEMENT_SSL_KEY_FILE_LOCATION | user management private key for inbound connections location |
| USER_MANAGEMENT_SSL_CA_FILE_LOCATION | user management certificate of certificate authority root  |
| WAITING_FOR_INSTANCE_STATE_CHANGE_RETRIES | Number of retries to wait for instance state change. Default value is 600. There's one second sleep between each check iteration. |
