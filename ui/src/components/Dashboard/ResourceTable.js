import React, { Fragment, useEffect, useState } from "react";
// import { useSelector } from 'react-redux'
import { connect } from "react-redux";
import PropTypes from 'prop-types';
import numeral from 'numeral';
import MUIDataTable from "mui-datatables";
import TextUtils from "utils/Text";
import TagsDialog from "../Dialog/Tags";
import { ResourcesService } from "services/resources.service";

const ResourceTable = ({ 
  filters,
  currentResource, 
  currentExecution, 
}) => {

  const [headers, setHeaders] = useState([]);
  const [rows, setRows] = useState([]);

  const tableOptions =  {
    selectableRows: false,
    responsive: "stacked",
  };

  const getRowRender = (key) => {
    let renderr = false;
    switch(key) {
      case "PricePerMonth":
      case "TotalSpendPrice":
        renderr =  (data) => (<span>{numeral(data).format('$ 0,0[.]00')}</span>)
      break;
      case "PricePerHour":
        renderr =  (data) => (<span>{numeral(data).format('$ 0,0[.]000')}</span>)
      break;
      case "Tag":
        renderr =  (data) => ( <TagsDialog tags={data} />)
      break;
      default:
        renderr =  (data) => (<span>{data}</span>)
    }
    return renderr;
  };

  const getHeaderRow = (exampleRow) => {
    const exclude = ['TotalSpendPrice']
    const keys = Object.keys(exampleRow).reduce((filtered, headerKey) => {
      if (exclude.indexOf(headerKey) === -1) {
        const header = {
          name: headerKey,
          label: TextUtils.CamelCaseToTitleCase(headerKey),
          options: {
            customBodyRender: getRowRender(headerKey)
          }
        }
         filtered.push(header);
      }
      return filtered;
    }, []);
    return keys;
  }

  const getData =  () => {
    if (!currentResource) {
      return currentResource;
    }
    ResourcesService.GetContent(currentResource, currentExecution, filters).then(responseData => {
      if (!responseData) {
        setHeaders([]);
        setRows([]);
        return false;
      }
      const headers = getHeaderRow(responseData[0].Data);
      const rows = responseData.map(row => row.Data)
      setHeaders(headers);
      setRows(rows);
    });
  }

  
  useEffect(() => {
    if (!currentExecution) {
      return;
    }
    getData();
  }, [currentExecution, currentResource, filters]);

  return (
    <Fragment>
      <div id="resourcewrap">
        <MUIDataTable
              data={rows}
              columns={headers}
              options={tableOptions}
              />
      </div>
    </Fragment>
    
  );
}


ResourceTable.defaultProps = {};
ResourceTable.propTypes = {
  currentExecution: PropTypes.string,
  currentResource: PropTypes.string,
  filters: PropTypes.array,
};

const mapStateToProps = state => ({
  currentExecution: state.executions.current,
  currentResource: state.resources.currentResource,
  filters: state.filters.filters,
});
const mapDispatchToProps = (dispatch) => ({
});

export default connect(mapStateToProps, mapDispatchToProps)(ResourceTable);