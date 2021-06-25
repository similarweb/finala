import React, { useState } from "react";
import { CSVLink } from "react-csv";
import { ResourcesService } from "services/resources.service";
import { connect } from "react-redux";
import PropTypes from "prop-types";

const CSVDownloadButton = ({ currentExecution, filters }) => {
  const [data, setData] = useState([]);
  const [csvLinkEl, setCsvLinkEl] = useState(React.createRef());

  //Testdata to Be ignored

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
        type="button"
        value="Download Current Data as CSV Report"
        onClick={downloadReport}
      />
      <CSVLink
        data={data}
        filename={"WIP Put ExecutionId here.csv"}
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
