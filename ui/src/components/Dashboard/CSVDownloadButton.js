import React from "react";
import { CSVLink } from "react-csv";
import { ResourcesService } from "services/resources.service";
import { connect } from "react-redux";
import { PropTypes } from "@material-ui/core";
import { useState } from "react";

const CSVDownloadButton = ({ executionId, filters }) => {
  const [data, setData] = useState([]);
  //Testdata to Be ignored
  const headers = [
    { label: "First Name", key: "firstName" },
    { label: "Last Name", key: "lastName" },
    { label: "Email", key: "email" },
    { label: "Age", key: "age" },
  ];

  const csvReport = {
    data: data,
    headers: headers,
    filename: "Clue_Mediator_Report.csv",
  };

  //Testdata to Be ignored

  const downloadReport = async () => {
    setData(await ResourcesService.GetReport(executionId, filters)); //tobefixed
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
        headers={headers}
        filename={"WIP Put ExecutionId here.csv"}
      />
    </div>
  );
};

//Todo
// -> Get the Execution ID and pass it do downloadReport

CSVDownloadButton.propTypes = {
  executionId: PropTypes.string,
  filters: PropTypes.array,
};

const mapStateToProps = (state) => ({
  executionId: state.executions.current,
  filters: state.filters.filters,
});

export default connect(mapStateToProps)(CSVDownloadButton);
