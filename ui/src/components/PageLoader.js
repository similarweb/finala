import React from "react";
import { Grid, CircularProgress, makeStyles } from "@material-ui/core";

const useStyles = makeStyles(() => ({
  root: {
    minHeight: "80vh",
    textAlign: "center",
  },
}));

const PageLoader = () => {
  const classes = useStyles();
  return (
    <Grid
      container
      spacing={0}
      direction="column"
      alignItems="center"
      justify="center"
      className={classes.root}
    >
      <Grid item xs={10}>
        <CircularProgress disableShrink size={80} />
      </Grid>
    </Grid>
  );
};

PageLoader.propTypes = {};
PageLoader.defaultProps = {};

export default PageLoader;
