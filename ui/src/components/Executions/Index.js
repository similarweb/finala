import React, { Fragment } from "react";
import Moment from "moment";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import { setHistory } from "../../utils/History";
import { ucfirstDirective } from "../../utils/Title";
import { Select, MenuItem } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";

const useStyles = makeStyles(() => ({
  selector: {
    marginTop: "10px",
    width: "100%",
    textAlign: "center",
    fontWeight: "bold",
    "& .MuiOutlinedInput-notchedOutline": {
      border: "0",
    },
    backgroundColor: "#d5dee6",
    borderColor: "#d5dee6",
    color: "rgba(0, 0, 0, 0.87)",
    border: "0",
    maxWidth: "320px",
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

  /**
   *
   * @param {string} executionId id of execution
   * update the current execution and set the url history
   */
  const updateCurrentExecution = (executionId) => {
    setCurrentExecution(executionId);
    setHistory({ executionId });
  };

  return (
    <Fragment>
      <Select
        className={classes.selector}
        variant="outlined"
        value={currentExecution}
        onChange={(event) => updateCurrentExecution(event.target.value)}
      >
        {executions.map((execution, i) => (
          <MenuItem key={i} value={execution.ID}>
            {ucfirstDirective(execution.Name)}{" "}
            {Moment(execution.Time).format("YYYY-MM-DD HH:mm")}
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
