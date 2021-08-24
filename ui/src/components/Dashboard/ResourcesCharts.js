import React from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import { Fragment } from "react";
import ResourcesChart from "./ResourcesChart";

/**
 * @param {accounts} array Accounts of current execution
 * @param {filters} array Filters list
 */
const ResourcesCharts = ({ accounts, filters }) => {
  let selectedAccountIds = filters
    .filter((filter) => filter.type === "account")
    .map((filter) => filter.id.split(":")[1]);
  if (selectedAccountIds.length === 0) {
    selectedAccountIds = accounts.map((account) => account.ID);
  }
  let resourcesCharts = selectedAccountIds.map((accountID) => (
    <ResourcesChart key={accountID} account={accountID} />
  ));
  resourcesCharts = [<ResourcesChart key={1} />, ...resourcesCharts];
  return <Fragment>{resourcesCharts}</Fragment>;
};

ResourcesCharts.defaultProps = {};
ResourcesCharts.propTypes = {
  filters: PropTypes.array,
  accounts: PropTypes.array,
};

const mapStateToProps = (state) => ({
  filters: state.filters.filters,
  accounts: state.accounts.accounts,
});

export default connect(mapStateToProps)(ResourcesCharts);
