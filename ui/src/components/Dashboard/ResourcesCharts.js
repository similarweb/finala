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
  const selectedAccountIds = filters
    .filter((filter) => filter.type === "account")
    .map((filter) => filter.id.split(":")[1]);
  let selectedAccounts;
  if (selectedAccountIds.length === 0) {
    selectedAccounts = accounts;
  } else {
    selectedAccounts = accounts.filter((account) =>
      selectedAccountIds.includes(account.ID)
    );
  }
  let resourcesCharts = selectedAccounts.map((account) => (
    <ResourcesChart key={account.ID} account={account} />
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
