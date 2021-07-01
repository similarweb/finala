import React, { useState } from "react";
import { CSVLink } from "react-csv";
import { ResourcesService } from "services/resources.service";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import { makeStyles } from "@material-ui/core/styles";
import { Button } from "@material-ui/core/Button"

const useStyles = makeStyles(() => ({
  myButton: {
    backgroundColor: "#d5dee6",
    borderColor: "#d5dee6",
    color: "rgba(0, 0, 0, 0.87)",
    border: "0",
    textAlign: "center",
    fontWeight: "bold",
    width: "100%",
    maxWidth: "320px",
    float: "right",
    borderRadius: "4px",
  },
}));

const CSVDownloadButton = ({ currentExecution, filters }) => {
  const [data, setData] = useState([]);
  const [csvLinkEl, setCsvLinkEl] = useState(React.createRef());
  const classes = useStyles();

  const downloadReport = async () => {
    const tempData = await ResourcesService.GetReport(
      currentExecution,
      filters
    ).catch(() => false);

    if (tempData) {
      setData(tempData);
    }
    csvLinkEl.current.link.click();
  };

  return (
    <div>
      <input
        className={classes.myButton}
        type="button"
        value="Download Current Data as CSV Report"
        onClick={downloadReport}
      />
      <Button
        variant="contained"
        size="small"
        value="Download Current Data as CSV Report"
        onClick={downloadReport}
      />
      <CSVLink
        data={data}
        filename={currentExecution + ".csv"}
        ref={csvLinkEl}
      />
    </div>
  );
};

CSVDownloadButton.defaultProps = {};
CSVDownloadButton.propTypes = {
  currentExecution: PropTypes.string,
  filters: PropTypes.array,
};

const mapStateToProps = (state) => ({
  currentExecution: state.executions.current,
  filters: state.filters.filters,
});

export default connect(mapStateToProps)(CSVDownloadButton);
