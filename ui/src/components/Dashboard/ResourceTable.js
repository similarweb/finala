import React, { Fragment, useEffect, useState } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import numeral from "numeral";
import MUIDataTable from "mui-datatables";
import TextUtils from "utils/Text";
import TagsDialog from "../Dialog/Tags";
import { ResourcesService } from "services/resources.service";
import ReportProblemIcon from "@material-ui/icons/ReportProblem";

import {
  makeStyles,
  Card,
  CardContent,
  LinearProgress,
} from "@material-ui/core";

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
  noDataTitle: {
    textAlign: "center",
    fontWeight: "bold",
    margin: "5px",
    fontSize: "14px",
  },
  AlertIcon: {
    fontSize: "56px",
    color: "red",
  },
  progress: {
    margin: "30px",
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
  const [isLoading, setIsLoading] = useState(true);

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
    setIsLoading(true);
    ResourcesService.GetContent(currentResource, currentExecution, filters)
      .then((responseData) => {
        setIsLoading(false);
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
        setIsLoading(false);
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
      setHasError(false);
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
      setHasError(false);
      getData();
    }

    // unmount, clear timers
    return () => {
      clearTimeout(fetchTimeout);
    };
  }, [resources, currentResource, filters, currentExecution]);

  return (
    <Fragment>
      {!hasError && isLoading && (
        <Card className={classes.Card}>
          <CardContent className={classes.CardContent}>
            <div className={classes.noDataTitle}>
              <LinearProgress className={classes.progress} />
            </div>
          </CardContent>
        </Card>
      )}

      {!isLoading && (hasError || !rows.length) && (
        <Card className={classes.Card}>
          <CardContent className={classes.CardContent}>
            {(hasError || !rows.length) && !isLoading && (
              <ReportProblemIcon className={classes.AlertIcon} />
            )}

            {hasError && (
              <h3>
                {
                  " Finala couldn't scan the selected resource, please check system logs "
                }
              </h3>
            )}

            {!hasError && !rows.length && (
              <div className={classes.noDataTitle}>
                <h3>No data found.</h3>
              </div>
            )}

            {errorMessage && <h4>{errorMessage}</h4>}
          </CardContent>
        </Card>
      )}

      {!hasError && rows.length > 0 && !isLoading && (
        <div id="resourcewrap">
          <MUIDataTable data={rows} columns={headers} options={tableOptions} />
        </div>
      )}
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
