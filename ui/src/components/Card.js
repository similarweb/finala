import React from "react";
import PropTypes from 'prop-types';
import { Col } from "react-bootstrap";
import SVG from 'react-inlinesvg';
import TextUtils from "utils/Text"

export default class Card extends React.Component {

  static propTypes = {    
    /**
     * List of all un-usage resources
     */
    Title: PropTypes.string, 

    Value: PropTypes.any, 

    Icon: PropTypes.any, 
  };  

  /**
  * Component render
  */    
  render() {
    return (
      <Col lg={3} md={6} sm={6}>
        <div className="card">
          <div className="card-body text-center">
            <SVG src={this.props.Icon} className="card-icon"/>
              <div className="content">
                  <p className="text-muted mt-2 mb-0">{TextUtils.CapitalizeWords(this.props.Title)}</p>
                  <p className="text-primary text-24 line-height-1 mb-2">{this.props.Value}</p>
              </div>
          </div>
        </div>
      </Col>
    );
  }
}
