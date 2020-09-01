import React, { Fragment, useEffect, useState } from "react";
import { connect } from "react-redux";
import PropTypes from "prop-types";
import numeral from "numeral";
import MUIDataTable from "mui-datatables";
import TextUtils from "utils/Text";
import TagsDialog from "../Dialog/Tags";
import ReportProblemIcon from "@material-ui/icons/ReportProblem";
import { getHistory } from "../../utils/History";
import { useTableFilters } from "../../Hooks/TableHooks";

import {
  makeStyles,
  Card,
  CardContent,
  LinearProgress,
} from "@material-ui/core";

import Moment from "moment";

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
 * @param  {array} {resources  Resources List
 * @param  {string} currentResource  Current Selected Resource
 * @param  {array} currentResourceData  Current Selected Resource data
 * @param  {bool} isResourceTableLoading  isLoading indicator for table}
 */
const ResourceTable = ({
  resources,
  currentResource,
  currentResourceData,
  isResourceTableLoading,
}) => {
  const [headers, setHeaders] = useState([]);
  const [errorMessage, setErrorMessage] = useState(false);
  const [hasError, setHasError] = useState(false);
  const classes = useStyles();
  const [setTableFilters] = useTableFilters({});
  const [tableOptions, setTableOptions] = useState({});

  // setting table configuration on first load
  useEffect(() => {
    setTableOptions({
      page: parseInt(getHistory("page", 0)),
      searchText: getHistory("search", ""),
      sortOrder: {
        name: getHistory("sortColumn", ""),
        direction: getHistory("direction", "desc"),
      },
      selectableRows: "none",
      responsive: "standard",
    });
  }, []);

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
        renderr = (data) => <span>{`${data}`}</span>;
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
   * Detect resource data changed
   */
  useEffect(() => {
    let headers = [];
    if (currentResourceData.length) {
      headers = getHeaderRow(currentResourceData[0]);
    }

    setHeaders(headers);
  }, [currentResourceData]);

  /**
   * Detect if we have an error
   */
  useEffect(() => {
    if (!currentResource) {
      return;
    }
    const resourceInfo = resources[currentResource];
    if (resourceInfo && resourceInfo.Status === 1) {
      setHasError(true);
      setErrorMessage(resourceInfo.ErrorMessage);
      return;
    } else {
      setHasError(false);
    }
  }, [currentResource, resources]);

  return (
    <Fragment>
      {!hasError && isResourceTableLoading && (
        <Card className={classes.Card}>
          <CardContent className={classes.CardContent}>
            <div className={classes.noDataTitle}>
              <LinearProgress className={classes.progress} />
            </div>
          </CardContent>
        </Card>
      )}

      {!isResourceTableLoading && (hasError || !currentResourceData.length) && (
        <Card className={classes.Card}>
          <CardContent className={classes.CardContent}>
            {(hasError || !currentResourceData.length) &&
              !isResourceTableLoading && (
                <ReportProblemIcon className={classes.AlertIcon} />
              )}

            {hasError && (
              <h3>
                {
                  " Finala couldn't scan the selected resource, please check system logs "
                }
              </h3>
            )}

            {!isResourceTableLoading &&
              !hasError &&
              !currentResourceData.length &&
              !headers.length && (
                <div className={classes.noDataTitle}>
                  <h3>No data found.</h3>
                </div>
              )}

            {errorMessage && <h4>{errorMessage}</h4>}
          </CardContent>
        </Card>
      )}

      {!hasError && currentResourceData.length > 0 && !isResourceTableLoading && (
        <div id="resourcewrap">
          <MUIDataTable
            data={currentResourceData}
            columns={headers}
            options={Object.assign(tableOptions, {
              onSearchChange: (searchText) => {
                setTableFilters([
                  {
                    key: "search",
                    value: searchText ? searchText : "",
                  },
                ]);
              },
              onColumnSortChange: (changedColumn, direction) => {
                setTableFilters([
                  { key: "sortColumn", value: changedColumn },
                  { key: "direction", value: direction },
                ]);
              },
              onChangePage: (currentPage) => {
                setTableFilters([{ key: "page", value: currentPage }]);
              },
              onChangeRowsPerPage: (numberOfRows) => {
                setTableFilters([{ key: "rows", value: numberOfRows }]);
              },
              downloadOptions: {
                filename: `${currentResource}.csv`,
              },
            })}
          />
        </div>
      )}
    </Fragment>
  );
};

ResourceTable.defaultProps = {};
ResourceTable.propTypes = {
  currentResource: PropTypes.string,
  resources: PropTypes.object,
  currentResourceData: PropTypes.array,
  isResourceTableLoading: PropTypes.bool,
};

const mapStateToProps = (state) => ({
  resources: state.resources.resources,
  currentResourceData: state.resources.currentResourceData,
  currentResource: state.resources.currentResource,
  isResourceTableLoading: state.resources.isResourceTableLoading,
});
const mapDispatchToProps = () => ({});

export default connect(mapStateToProps, mapDispatchToProps)(ResourceTable);
