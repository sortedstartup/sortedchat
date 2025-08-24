# Step 1 prepaase json
{
    clientid
    redirect_uri
    respons_type
    scope
    access_type
    state   
}

# Step 2 Reirect to Google oAuth 2.0
```
https://accounts.google.com/o/oauth2/v2/auth?
 ///scope=https%3A//www.googleapis.com/auth/drive.metadata.readonly%20https%3A//www.googleapis.com/auth/calendar.readonly&
 // for login only need this ->
 scope=openid%20email&
 access_type=offline&
 include_granted_scopes=true&
 response_type=code&
 state=state_parameter_passthrough_value&
 redirect_uri=https%3A//oauth2.example.com/code&
 client_id=client_id
 ```

# Step 3 Google shows UI and handles it
google generate a temporary auth code

# Step 4 google -> http://backend/callback?auth=code
we get auth code

# Step 5 backend -> (get auth token from google) google POST https://oauth2.googleapis.com/token

```
POST /token HTTP/1.1
Host: oauth2.googleapis.com
Content-Type: application/x-www-form-urlencoded

code=4/P7q7W91a-oMsCeLvIaQm6bTrgtp7&
client_id=your_client_id&
client_secret=your_client_secret&
redirect_uri=https%3A//oauth2.example.com/code&
grant_type=authorization_code
```

Google gives -
<---
{
  "access_token": "1/fFAGRNJru1FTz70BzhT3Zg",
  "expires_in": 3920,
  "id_token": "A JWT signed by google"
  "token_type": "Bearer",
// for login scope: "..smaller.."
  //"scope": "https://www.googleapis.com/auth/drive.metadata.readonly https://www.googleapis.com/auth/calendar.readonly",
  "refresh_token": "1//xEoDL4iW3cxlI7yDbSRFYNG01kVKM2C-259HOF2aQbI"
}

# Step 6. check scopes

# Step 7 decode id_token JWT and get user info
https://developers.google.com/identity/openid-connect/openid-connect#exchangecode

{
  "iss": "https://accounts.google.com",
  "azp": "1234987819200.apps.googleusercontent.com",
  "aud": "1234987819200.apps.googleusercontent.com",
  "sub": "10769150350006150715113082367",
  "at_hash": "HK6E_P6Dh8Y93mRNtsDB1Q",
  "hd": "example.com",
  "email": "jsmith@example.com",
  "email_verified": "true",
  "iat": 1353601026,
  "exp": 1353604926,
  "nonce": "0394852-3190485-2490358"
}

Ref:
https://developers.google.com/identity/openid-connect/openid-connect


# Step 8 Recommeed by google if user does not exist in db create it now

quote from google page
After obtaining user information from the ID token, you should query your app's user database. If the user already exists in your database, you should start an application session for that user if all login requirements are met by the Google API response.

If the user does not exist in your user database, you should redirect the user to your new-user sign-up flow. You may be able to auto-register the user based on the information you receive from Google, or at the very least you may be able to pre-populate many of the fields that you require on your registration form. In addition to the information in the ID token, you can get additional user profile information at our user profile endpoints.