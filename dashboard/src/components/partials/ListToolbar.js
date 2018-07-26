import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import CollapsingMenu from "/components/partials/CollapsingMenu";

const FillSpace = withStyles({
  root: {
    flex: "1 1 auto",
  },
})(({ classes }) => <div className={classes.root} />);

const styles = theme => ({
  search: {
    width: "100%",
    [theme.breakpoints.up("sm")]: {
      width: "50%",
    },
  },
});

class ListToolbar extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    renderSearch: PropTypes.node.isRequired,
    renderMenuItems: PropTypes.node.isRequired,
  };

  static defaultProps = {
    renderSearch: null,
  };

  render() {
    const { classes, renderSearch } = this.props;

    let search;
    if (renderSearch) {
      search = React.cloneElement(renderSearch, {
        className: classnames(classes.search, renderSearch.props.className),
      });
    }

    return (
      <React.Fragment>
        {search}
        <FillSpace />
        <CollapsingMenu breakpoint="sm">
          {this.props.renderMenuItems}
        </CollapsingMenu>
      </React.Fragment>
    );
  }
}

export default withStyles(styles)(ListToolbar);
