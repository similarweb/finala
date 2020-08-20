import React, { Fragment } from "react";
import { connect } from "react-redux";
import {
  makeStyles,
  Box,
  Card,
  CardContent,
  LinearProgress,
} from "@material-ui/core";

import Logo from "./Logo";

const useStyles = makeStyles(() => ({
  Root: {
    maxWidth: "600px",
    margin: "15% auto",
    textAlign: "center",
  },
  Card: {
    marginTop: "20px",
  },
  CardContent: {
    padding: "30px",
  },
}));

const NoData = () => {
  const classes = useStyles();
  return (
    <Fragment>
      <div className={classes.Root}>
        <Box mb={3}>
          <Logo />
          <Card className={classes.Card}>
            <CardContent className={classes.CardContent}>
              <h3>Waiting for the first collection of data for Finala</h3>
              <br />
              <LinearProgress />
            </CardContent>
          </Card>
        </Box>
      </div>
    </Fragment>
  );
};

NoData.defaultProps = {};
NoData.propTypes = {};

const mapStateToProps = () => ({});
const mapDispatchToProps = () => ({});

export default connect(mapStateToProps, mapDispatchToProps)(NoData);
