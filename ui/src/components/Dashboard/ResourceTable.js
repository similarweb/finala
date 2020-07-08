import React, { Fragment, useEffect, useState } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import numeral from "numeral";
import MUIDataTable from "mui-datatables";
import TextUtils from "utils/Text";
import TagsDialog from "../Dialog/Tags";
import { ResourcesService } from "services/resources.service";

/**
 * @param  {array} {filters  Filters List
 * @param  {func} currentResource  Current Selected Resource
 * @param  {func} currentExecution Current Selected Execution}
 */
const ResourceTable = ({ filters, currentResource, currentExecution }) => {
  const [headers, setHeaders] = useState([]);
  const [rows, setRows] = useState([]);

  const tableOptions = {
    selectableRows: "none",
    responsive: "standard",
  };

  /**
   * format table cell by type
   * @param {string} key TableCell key
   * @returns {func} render function to render cell
   */
  const getRowRender = (key) => {
    let renderr = false;
    switch (key) {
      case "PricePerMonth":
      case "TotalSpendPrice":
        renderr = (data) => <span>{numeral(data).format("$ 0,0[.]00")}</span>;
        break;
      case "PricePerHour":
        renderr = (data) => <span>{numeral(data).format("$ 0,0[.]000")}</span>;
        break;
      case "Tag":
        renderr = (data) => <TagsDialog tags={data} />;
        break;
      default:
        renderr = (data) => <span>{data}</span>;
    }
    return renderr;
  };

  /**
   * determines Table header keys
   * @param {object} exampleRow  sample row from data
   * @returns {array} Table header keys
   */
  const getHeaderRow = (exampleRow) => {
    const exclude = ["TotalSpendPrice"];
    const keys = Object.keys(exampleRow).reduce((filtered, headerKey) => {
      if (exclude.indexOf(headerKey) === -1) {
        const header = {
          name: headerKey,
          label: TextUtils.CamelCaseToTitleCase(headerKey),
          options: {
            customBodyRender: getRowRender(headerKey),
          },
        };
        filtered.push(header);
      }
      return filtered;
    }, []);
    return keys;
  };

  /**
   * fetch data for global selected resource
   */
  const getData = () => {
    if (!currentResource) {
      return currentResource;
    }
    ResourcesService.GetContent(
      currentResource,
      currentExecution,
      filters
    ).then((responseData) => {
      if (!responseData) {
        setHeaders([]);
        setRows([]);
        return false;
      }
      const headers = getHeaderRow(responseData[0].Data);
      const rows = responseData.map((row) => row.Data);
      setHeaders(headers);
      setRows(rows);
    });
  };

  /**
   * refetch data when state changes
   */
  useEffect(() => {
    if (!currentExecution) {
      return;
    }
    getData();
  }, [currentExecution, currentResource, filters]);

  return (
    <Fragment>
      <div id="resourcewrap">
        <MUIDataTable data={rows} columns={headers} options={tableOptions} />
      </div>
    </Fragment>
  );
};

ResourceTable.defaultProps = {};
ResourceTable.propTypes = {
  currentExecution: PropTypes.string,
  currentResource: PropTypes.string,
  filters: PropTypes.array,
};

const mapStateToProps = (state) => ({
  currentExecution: state.executions.current,
  currentResource: state.resources.currentResource,
  filters: state.filters.filters,
});
const mapDispatchToProps = () => ({});

export default connect(mapStateToProps, mapDispatchToProps)(ResourceTable);
