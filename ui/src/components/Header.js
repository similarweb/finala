import React from "react";
import { Navbar, Nav } from "react-bootstrap";
import { Link } from "react-router-dom";


/**
 * Application Header
 */
export default class Header extends React.Component {

  /**
  * Component render
  */    
  render() {
    return (
      <div id="header-login">
        <Navbar bg="light" variant="light">
          <Link className="navbar-brand" to={`/`}>Finala</Link>
          <Nav className="mr-auto">
            <Link to={`/`}>Dashboard</Link> 
          </Nav>
        </Navbar>
      </div>
    );
  }
}
