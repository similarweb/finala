import React, { Fragment } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import { Select, MenuItem } from "@material-ui/core";

import { makeStyles } from "@material-ui/core/styles";

const useStyles = makeStyles(() => ({
  selector: {
    marginTop: "10px",
    width: "100%",
    fontWeight: "bold",
    "& .MuiOutlinedInput-notchedOutline": {
      border: "0",
    },
    backgroundColor: "#d5dee6",
    borderColor: "#d5dee6",
    color: "rgba(0, 0, 0, 0.87)",
    border: "0",
    maxWidth: "290px",
  },
}));

/**
 * @param  {array} {executions  Executions List
 * @param  {string} currentExecution Global Execution Id
 * @param  {func} setCurrentExecution Update Current Execution}
 */
const ExecutionsIndex = ({
  executions,
  currentExecution,
  setCurrentExecution,
}) => {
  const classes = useStyles();
  return (
    <Fragment>
      <Select
        className={classes.selector}
        variant="outlined"
        value={currentExecution}
        onChange={(event) => setCurrentExecution(event.target.value)}
      >
        {executions.map((execution, i) => (
          <MenuItem key={i} value={execution.ID}>
            {execution.Name} {execution.Time}
          </MenuItem>
        ))}
      </Select>
    </Fragment>
  );
};

ExecutionsIndex.defaultProps = {};
ExecutionsIndex.propTypes = {
  currentExecution: PropTypes.string,
  executions: PropTypes.array,
  setCurrentExecution: PropTypes.func,
};

const mapStateToProps = (state) => ({
  currentExecution: state.executions.current,
  executions: state.executions.list,
});

const mapDispatchToProps = (dispatch) => ({
  setCurrentExecution: (id) => dispatch({ type: "EXECUTION_SELECTED", id }),
});

export default connect(mapStateToProps, mapDispatchToProps)(ExecutionsIndex);
