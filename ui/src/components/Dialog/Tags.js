import React from "react";
import PropTypes from "prop-types";
import Button from "@material-ui/core/Button";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogTitle from "@material-ui/core/DialogTitle";
import Link from "@material-ui/core/Link";

/**
 *
 * @param {string} tags - tags list
 */
const DialogTags = ({ tags }) => {
  const [open, setOpen] = React.useState(false);

  /**
   * Open dialog
   */
  const handleClickOpen = () => {
    setOpen(true);
  };

  /**
   * Close dialog
   */
  const handleClose = () => {
    setOpen(false);
  };

  return (
    <React.Fragment>
      <Link component="button" variant="body2" onClick={handleClickOpen}>
        Tags
      </Link>

      <Dialog
        open={open}
        onClose={handleClose}
        aria-labelledby="max-width-dialog-title"
      >
        <DialogTitle id="max-width-dialog-title">Tags</DialogTitle>
        <DialogContent>
          <DialogContentText></DialogContentText>
          <pre>{<div>{JSON.stringify(tags, null, 2)}</div>}</pre>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            Close
          </Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  );
};

DialogTags.propTypes = {
  tags: PropTypes.object,
};

DialogTags.defaultProps = {};

export default DialogTags;
