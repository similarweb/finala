import React from "react";
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import Link from '@material-ui/core/Link';
import { makeStyles } from '@material-ui/core/styles';
import {history} from 'configureStore'
import Box from '@material-ui/core/Box';

const useStyles = makeStyles((theme) => ({
  appBar: {
    zIndex: theme.zIndex.drawer + 1,
  },
  link: {
    color: "white"
  },
}));

/**
 * Application Header
 */
const Header = () => {

  const classes = useStyles();

  const goToHomepage = () => history.push("/");


 
  /**
  * Component render
  */    
    return (
      <AppBar position="fixed" className={classes.appBar}>
        <Toolbar>
          <Box component="span" mr={5}>
            <Typography variant="h6" noWrap>
              <Link href="#" className={classes.link} onClick={goToHomepage}>
                Finala
              </Link>
            </Typography>
          </Box>

          <Typography className={classes.root}>
            <Link href="#" onClick={goToHomepage} color="inherit">
              Dashboard
            </Link>
        </Typography>
        
        </Toolbar>
      </AppBar>
    );
}

Header.propTypes = {
};

Header.defaultProps = {

};

export default Header;