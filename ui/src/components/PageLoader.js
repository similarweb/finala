import React from 'react'
import CircularProgress from '@material-ui/core/CircularProgress';
import Grid from '@material-ui/core/Grid';


const PageLoader = () => {

  return (

    <Grid
      container
      spacing={0}
      direction="column"
      alignItems="center"
      justify="center"
      style={{ minHeight: '80vh', textAlign: "center" }}
    >
      <Grid item xs={10}>
        <CircularProgress disableShrink  size={80} />
      </Grid>   

    </Grid> 
  )
}

PageLoader.propTypes = {};
PageLoader.defaultProps = {};

export default PageLoader;