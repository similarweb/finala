import React from "react";
import PropTypes from "prop-types";
import { ConnectedRouter } from "connected-react-router";
import Routes from "./routes";
import { connect } from "react-redux";
import "./styles/index.scss";

// Main application class
class App extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <ConnectedRouter history={this.props.history} basename="">
        <Routes />
      </ConnectedRouter>
    );
  }
}

function mapStateToProps() {
  return {};
}

App.propTypes = {
  history: PropTypes.object,
};
export default connect(mapStateToProps)(App);
