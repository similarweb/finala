import React, { Fragment, useEffect, useState } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import numeral from "numeral";
import MUIDataTable from "mui-datatables";
import TextUtils from "utils/Text";
import TagsDialog from "../Dialog/Tags";
import { ResourcesService } from "services/resources.service";
import ReportProblemIcon from "@material-ui/icons/ReportProblem";

import { makeStyles, Card, CardContent } from "@material-ui/core";

import Moment from "moment";

let fetchTimeout = false;
let lastResource = false;

const useStyles = makeStyles(() => ({
  Card: {
    marginBottom: "20px",
  },
  CardContent: {
    padding: "30px",
    textAlign: "center",
  },
  AlertIcon: {
    fontSize: "56px",
    color: "red",
  },
}));

/**
 * @param  {array} {filters  Filters List
 * @param  {array} resources  Resources List
 * @param  {func} currentResource  Current Selected Resource
 * @param  {func} currentExecution Current Selected Execution}
 */
const ResourceTable = ({
  filters,
  resources,
  currentResource,
  currentExecution,
}) => {
  const [headers, setHeaders] = useState([]);
  const [rows, setRows] = useState([]);
  const [errorMessage, setErrorMessage] = useState(false);
  const [hasError, setHasError] = useState(false);

  const classes = useStyles();

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
      case "LaunchTime":
        renderr = (data) => (
          <span>{Moment(data).format("YYYY-MM-DD HH:mm")}</span>
        );
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
  const getHeaderRow = (row) => {
    const exclude = ["TotalSpendPrice"];
    const keys = Object.keys(row).reduce((filtered, headerKey) => {
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
    clearTimeout(fetchTimeout);
    if (!currentResource) {
      return currentResource;
    }
    ResourcesService.GetContent(currentResource, currentExecution, filters)
      .then((responseData) => {
        if (!responseData) {
          setHeaders([]);
          setRows([]);
          return false;
        }
        const headers = getHeaderRow(responseData[0].Data);
        const rows = responseData.map((row) => row.Data);
        const resourceInfo = resources[currentResource];

        setHeaders(headers);
        setRows(rows);

        if (resourceInfo && resourceInfo.Status === 0) {
          fetchTimeout = setTimeout(getData, 5000);
        } else {
          clearTimeout(fetchTimeout);
        }
      })
      .catch(() => {
        fetchTimeout = setTimeout(getData, 5000);
      });
  };

  useEffect(() => {
    if (!currentExecution || !currentResource) {
      return;
    }
    let shouldRefreshData = false;
    const resourceInfo = resources[currentResource];

    // resource not exists in selected execution
    if (!resourceInfo) {
      setRows([]);
    }

    // keep scanning
    if (resourceInfo && resourceInfo.Status === 0) {
      shouldRefreshData = true;
    }

    // check for error status
    if (resourceInfo && resourceInfo.Status === 1) {
      setHasError(true);
      setErrorMessage(resourceInfo.ErrorMessage);
    }

    // new resource selected
    if (JSON.stringify(lastResource) !== JSON.stringify(resourceInfo)) {
      lastResource = resourceInfo;
      shouldRefreshData = true;
    }

    // inital refresh
    if (!rows) {
      shouldRefreshData = true;
    }
    if (shouldRefreshData) {
      getData();
    }

    // unmount, clear timers
    return () => {
      clearTimeout(fetchTimeout);
    };
  }, [resources, currentResource, filters, currentExecution]);

  return (
    <Fragment>
      {hasError && (
        <Card className={classes.Card}>
          <CardContent className={classes.CardContent}>
            <ReportProblemIcon className={classes.AlertIcon} />
            <h3>
              {
                " Finala couldn't scan the selected resource, please check system logs "
              }
            </h3>
            {errorMessage && <h4>{errorMessage}</h4>}
          </CardContent>
        </Card>
      )}

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
  resources: PropTypes.object,
  filters: PropTypes.array,
};

const mapStateToProps = (state) => ({
  currentExecution: state.executions.current,
  resources: state.resources.resources,
  currentResource: state.resources.currentResource,
  filters: state.filters.filters,
});
const mapDispatchToProps = () => ({});

export default connect(mapStateToProps, mapDispatchToProps)(ResourceTable);
