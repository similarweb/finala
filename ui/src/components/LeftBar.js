import React from "react";
import PropTypes from 'prop-types';
import {connect} from 'react-redux';
import { Link } from "react-router-dom";
import { Col } from "react-bootstrap";
import TextUtils from "utils/Text";
import LoaderDots from "components/Loader/Dots";
import SVG from 'react-inlinesvg';
import XSVG from "styles/icons/x.svg"

@connect(state => ({
  resources: state.resources,
}))
/**
 * Application left bar menu
 */
export default class LeftBar extends React.Component {

  static propTypes = {    
    /**
     * List of all un-usage resources
     */
    resources : PropTypes.object, 
  };  

  /**
  * Component render
  */    
  render() {
    return (
      <div className="row flex-column" id="left-bar">
        <Col sm={12}>
            <ul>
              <li className="title">Resources</li>
              {Object.keys(this.props.resources).map((resource) =>
                <li className="option" key={`resource ${resource}`}>
                  
                  {this.props.resources[resource].Status == 1 && <span title={this.props.resources[resource].Description}><SVG src={XSVG} className="failed pull-left"/></span>}
                  <Link className="pull-left" to={`/resource/${resource}`}>{TextUtils.ParseName(resource)} ({this.props.resources[resource].ResourceCount}) </Link> 
                  {this.props.resources[resource].Status == 0 && <div className="pull-left"><LoaderDots /></div>}
                </li>
              )}
            </ul>
        </Col>
      </div>
    );
  }
}
