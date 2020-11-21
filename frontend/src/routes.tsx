import React from 'react'
import {
    BrowserRouter as Router,
    Switch,
    Route,
  } from 'react-router-dom';
import { Home } from './pages/Home'
import { PrivacyPolicy } from './pages/PrivacyPolicy'

export const Routes = () => {
    return (
        <Router>
            <div>
                {/* A <Switch> looks through its children <Route>s and
                    renders the first one that matches the current URL. */}
                <Switch>
                <Route path="/about">
                    <About />
                </Route>
                <Route path="/users">
                    <Users />
                </Route>
                <Route path="/privacy-policy">
                    <PrivacyPolicy />
                </Route>
                <Route path="/">
                    <Home />
                </Route>
                </Switch>
            </div>
        </Router>
    );
}

function About() {
    return <h2>About</h2>;
}

function Users() {
    return <h2>Users</h2>;
}