import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/Button";
import Disclosure from "@material-ui/icons/MoreVert";

const styles = theme => ({
  root: {
    minWidth: 0,
    paddingLeft: theme.spacing.unit / 2,
    paddingRight: theme.spacing.unit / 2,
  },
});

class VerticalDisclosureButton extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
  };

  static defaultProps = {
    className: "",
  };

  render() {
    const { className: classNameProp, classes, ...props } = this.props;
    const className = classnames(classes.root, classNameProp);

    return (
      <Button className={className} {...props}>
        <Disclosure />
      </Button>
    );
  }
}

export default withStyles(styles)(VerticalDisclosureButton);
