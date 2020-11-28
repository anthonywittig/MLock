import React from 'react';
import {
    GoogleLogin,
    GoogleLoginResponse,
    GoogleLoginResponseOffline, 
} from 'react-google-login';

const responseGoogleSuccess = (response: GoogleLoginResponse | GoogleLoginResponseOffline) => {
    console.log(response);
}

const responseGoogleFailure = (response: any) => {
    console.log(response);
}

export const SignIn = () => {
    return (<div>
        <h2>Sign In</h2>
        <br />
        <GoogleLogin
            clientId={process.env.REACT_APP_GOOGLE_SIGNIN_CLIENT_ID || ""}
            buttonText="Login"
            onSuccess={responseGoogleSuccess}
            onFailure={responseGoogleFailure}
            cookiePolicy={'single_host_origin'}
        />
        <div className="g-signin2" data-onsuccess="onSignIn"></div>
    </div>);
}