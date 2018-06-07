import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Checkbox from "@material-ui/core/Checkbox";
import { withStyles } from "@material-ui/core/styles";
import { capitalize } from "lodash";

import Typography from "@material-ui/core/Typography";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemText from "@material-ui/core/ListItemText";

import {
  TableListHeader,
  TableListSelect,
  TableListButton as Button,
} from "/components/TableList";

import StatusMenu from "/components/partials/StatusMenu";

const styles = theme => ({
  headerButton: {
    marginLeft: theme.spacing.unit / 2,
    "&:first-child": {
      marginLeft: theme.spacing.unit,
    },
  },
  filterActions: {
    display: "none",
    [theme.breakpoints.up("sm")]: {
      display: "flex",
    },
  },
  // Remove padding from button container
  checkbox: {
    marginLeft: -11,
    color: theme.palette.primary.contrastText,
  },
  grow: {
    flex: "1 1 auto",
  },
});

class EntitiesListHeader extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    onClickSelect: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
    onClickResolve: PropTypes.func.isRequired,
    selectedCount: PropTypes.number.isRequired,
    environment: PropTypes.shape({
      checks: PropTypes.object,
      entities: PropTypes.object,
    }),
    onQueryChange: PropTypes.func.isRequired,
  };

  static defaultProps = {
    environment: null,
  };

  static fragments = {
    environment: gql`
      fragment EventsListHeader_environment on Environment {
        checks(limit: 1000) {
          nodes {
            name
          }
        }

        entities(limit: 1000) {
          nodes {
            name
          }
        }
      }
    `,
  };

  requeryEntity = newValue => {
    this.props.onQueryChange({ filter: `Entity.ID == '${newValue}'` });
  };

  requeryCheck = newValue => {
    this.props.onQueryChange({ filter: `Check.Name == '${newValue}'` });
  };

  requeryStatus = newValue => {
    if (Array.isArray(newValue)) {
      if (newValue.length === 1) {
        this.props.onQueryChange({ filter: `Check.Status == ${newValue}` });
      } else {
        const val = newValue.join(",");
        this.props.onQueryChange({ filter: `Check.Status IN (${val})` });
      }
    } else {
      this.props.onQueryChange(query => query.delete("filter"));
    }
  };

  requerySort = newValue => {
    this.props.onQueryChange({ order: newValue });
  };

  render() {
    const {
      classes,
      selectedCount,
      onClickSelect,
      onClickSilence,
      onClickResolve,
      environment,
    } = this.props;

    const entityNames = environment
      ? environment.entities.nodes.map(node => node.name)
      : [];

    const checkNames = [
      ...(environment ? environment.checks.nodes.map(node => node.name) : []),
      "keepalive",
    ];

    return (
      <TableListHeader sticky active={selectedCount > 0}>
        <Checkbox
          component="button"
          className={classes.checkbox}
          onClick={onClickSelect}
          checked={false}
          indeterminate={selectedCount > 0}
        />
        {selectedCount > 0 && <div>{selectedCount} Selected</div>}
        <div className={classes.grow} />
        {selectedCount > 0 && (
          <div>
            <Button className={classes.headerButton} onClick={onClickSilence}>
              <Typography variant="button">Silence</Typography>
            </Button>
            <Button className={classes.headerButton} onClick={onClickResolve}>
              <Typography variant="button">Resolve</Typography>
            </Button>
          </div>
        )}
        {selectedCount > 0 && (
          <div className={classes.filterActions}>
            <TableListSelect
              className={classes.headerButton}
              label="Entity"
              onChange={this.requeryEntity}
            >
              {entityNames.map(name => (
                <MenuItem key={name} value={name}>
                  <ListItemText primary={name} />
                </MenuItem>
              ))}
            </TableListSelect>
            <TableListSelect
              className={classes.headerButton}
              label="Check"
              onChange={this.requeryCheck}
            >
              {checkNames.map(name => (
                <MenuItem key={name} value={name}>
                  <ListItemText primary={name} />
                </MenuItem>
              ))}
            </TableListSelect>
            <StatusMenu
              className={classes.headerButton}
              onChange={this.requeryStatus}
            />
            <TableListSelect
              className={classes.headerButton}
              label="Sort"
              onChange={this.requerySort}
            >
              {["SEVERITY", "NEWEST", "OLDEST"].map(name => (
                <MenuItem key={name} value={name}>
                  <ListItemText primary={capitalize(name)} />
                </MenuItem>
              ))}
            </TableListSelect>
          </div>
        )}
      </TableListHeader>
    );
  }
}

export default withStyles(styles)(EntitiesListHeader);
