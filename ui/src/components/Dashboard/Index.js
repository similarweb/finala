import React from "react";
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import { Row } from "react-bootstrap";
import Card from "components/Card";
import CloudSVG from "styles/icons/cloud.svg"
import MoneyBagSVG from "styles/icons/money-bag.svg"
import NumberUtils from "utils/Number"
import TextUtils from "utils/Text"

/**
 * Dashboard page 
 */
@connect(state => ({
  resources: state.resources,
}))
export default class Dashboard extends React.Component {
  
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
      <div className="">
        <h1>Dashboard</h1>
        {Object.keys(this.props.resources).map((resource) =>
          <div key={`summary-${resource}`} className="cards">
            <h3>{TextUtils.ParseName(resource)}</h3>
            <Row>
              <Card Title="unused count" Value={this.props.resources[resource].ResourceCount} Icon={CloudSVG}/>  
              <Card Title="total spent" Value={`$${NumberUtils.Format(this.props.resources[resource].TotalSpent, 2)}`} Icon={MoneyBagSVG}/>  
            </Row>
          </div>
        )}
      </div>
    );
  }
}
