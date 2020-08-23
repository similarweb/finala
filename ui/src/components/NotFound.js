import React from "react";
import { connect } from "react-redux";
import { history } from "configureStore";

@connect()
/**
 * Route not found
 */
export default class NotFound extends React.Component {
  /**
   * When component mount redirect to root route
   */
  componentDidMount() {
    history.push("/");
  }

  /**
   * Component render
   */

  render() {
    return "";
  }
}
