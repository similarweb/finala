import React, { Fragment } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import colors from "./colors.json";
import { makeStyles } from "@material-ui/core/styles";
import { Box, Chip } from "@material-ui/core";
import { setHistory } from "../../utils/History";

const useStyles = makeStyles(() => ({
  title: {
    fontFamily: "MuseoModerno",
  },
  resource_chips: {
    fontWeight: "bold",
    fontFamily: "Arial !important",
    margin: "5px",
    borderRadius: "1px",
    backgroundColor: "#ffffff",
    borderLeft: "5px solid #ffffff",
    fontSize: "14px",
  },
}));

/**
 * @param  {array} accounts  Accounts List
 * @param  {array} filters  Filters List
 * @param  {func} addFilter Add filter to  filters list
 */
const AccountsList = ({ accounts, filters, addFilter }) => {
  const classes = useStyles();

  const accountsList = Object.values(accounts).map((account) => {
    account.title = `${account.Name}(${account.ID})`;
    return account;
  });

  /**
   *
   * @param {object} account add selected account
   */
  const setSelectedAccount = (account) => {
    const filter = {
      title: `Account:${account.title}`,
      id: `account:${account.ID}`,
      type: "account",
    };

    addFilter(filter);

    setHistory({
      filters: filters,
    });
  };

  return (
    <Fragment>
      {accountsList.length > 0 && (
        <Box mb={3}>
          <h4 className={classes.title}>Accounts:</h4>
          {accountsList.map((account, i) => (
            <Chip
              className={classes.resource_chips}
              style={{ borderLeftColor: colors[i].hex }}
              onClick={() => setSelectedAccount(account)}
              ma={2}
              label={account.title}
              key={i}
            />
          ))}
        </Box>
      )}
    </Fragment>
  );
};

AccountsList.defaultProps = {};
AccountsList.propTypes = {
  accounts: PropTypes.object,
  filters: PropTypes.array,
  addFilter: PropTypes.func,
};

const mapStateToProps = (state) => ({
  accounts: state.accounts.accounts,
  filters: state.filters.filters,
});
const mapDispatchToProps = (dispatch) => ({
  addFilter: (data) => dispatch({ type: "ADD_FILTER", data }),
});

export default connect(mapStateToProps, mapDispatchToProps)(AccountsList);
