import React from "react";
import Button from "@material-ui/core/Button";
import Grid from "@material-ui/core/Grid";
import TextField from "@material-ui/core/TextField";
import { makeStyles } from "@material-ui/core/styles";
import PropTypes from "prop-types";
import { AuthService } from "../../services/authentication.service";
import { connect } from "react-redux";
import Logo from "../Logo";
import Box from "@material-ui/core/Box";
import { Fragment, useState } from "react";

const useStyles = makeStyles((theme) => ({
  container: {
    padding: theme.spacing(3),
  },
  box: {
    textAlign: "center",
    marginRight: "auto",
    marginLeft: "auto",
    maxWidth: "600px",
  },
  error: {
    color: "red",
  },
}));

/**
 * @param {boolean} setAuthRequired tells if authentication is needed
 */
const Login = ({ setAuthRequired }) => {
  const classes = useStyles();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  const onSubmit = async () => {
    setErrorMessage("");
    const ok = await AuthService.Auth(username, password).catch(() => false);
    if (ok) {
      setAuthRequired(false);
    } else {
      setErrorMessage("Login Failed");
    }
  };

  return (
    <Fragment>
      <Box className={classes.box}>
        <Logo />
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  onChange={(e) => setUsername(e.target.value)}
                  onKeyPress={(e) => {
                    if (e.key == "Enter") {
                      onSubmit();
                    }
                  }}
                  fullWidth
                  label="Username"
                  name="username"
                  size="small"
                  variant="outlined"
                  value={username}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  onChange={(e) => setPassword(e.target.value)}
                  onKeyPress={(e) => {
                    if (e.key == "Enter") {
                      onSubmit();
                    }
                  }}
                  fullWidth
                  label="Password"
                  name="password"
                  size="small"
                  type="password"
                  variant="outlined"
                  value={password}
                />
              </Grid>
            </Grid>
          </Grid>
          <Grid item xs={12}>
            <Button
              color="primary"
              fullWidth
              type="submit"
              variant="contained"
              onClick={onSubmit}
            >
              Log in
            </Button>
          </Grid>
          <Grid item xs={12} className={classes.error}>
            {errorMessage}
          </Grid>
        </Grid>
      </Box>
    </Fragment>
  );
};

Login.defaultProps = {};
Login.propTypes = {
  setAuthRequired: PropTypes.func,
};

const mapDispatchToProps = (dispatch) => ({
  setAuthRequired: (authRequired) =>
    dispatch({ type: "SET_AUTH_REQUIRED", authRequired }),
});

export default connect(null, mapDispatchToProps)(Login);
