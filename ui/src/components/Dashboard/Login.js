import React from 'react';
import Button from '@material-ui/core/Button';
import Container from '@material-ui/core/Container';
import Grid from '@material-ui/core/Grid';
import TextField from '@material-ui/core/TextField';
import { makeStyles } from '@material-ui/core/styles';
import {PropTypes} from "@material-ui/core";
import { AuthService } from "../services/authentification.service";
import { useForm } from 'react-hook-form';

const useStyles = makeStyles((theme) => ({
  container: {
    padding: theme.spacing(3),
  },
});

interface FormData {
  username: string;
  password: string;
}


const Login = ({
  authRequired,
  }) => {
  const attemptLogin = (username, password) => {
    if (AuthService.Auth( username,password) == 200){
      setAuthRequired(false)
    } else {
      //huh
    }
  };

  const { handleSubmit, register } = useForm<FormData>();

  const classes = useStyles();

  const onSubmit = handleSubmit((data) => {
    attemptLogin(data.username, data.password)
  });

  return (
    <Container className={classes.container} maxWidth="xs">
      <form onSubmit={onSubmit}>
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  inputRef={register}
                  label="username"
                  name="eusername"
                  size="small"
                  variant="outlined"
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  inputRef={register}
                  label="Password"
                  name="password"
                  size="small"
                  type="password"
                  variant="outlined"
                />
              </Grid>
            </Grid>
          </Grid>
          <Grid item xs={12}>
            <Button color="secondary" fullWidth type="submit" variant="contained">
              Log in
            </Button>
          </Grid>
        </Grid>
      </form>
    </Container>
  )
}

Login.defaultProps = {};
Login.propTypes = {
  authRequired: PropTypes.bool,
}

const mapStateToProps = (state) => ({
  authRequired: state.accounts.authRequired,
})

const mapDispatchToProps = (dispatch) => ({
  setAuthRequired: (authRequired) => dispatch({type: "AUTH_REQUIRED", authRequired}),
})
