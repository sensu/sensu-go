import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Button from "@material-ui/core/Button";
import { withStyles } from "@material-ui/core/styles";
import capitalize from "lodash/capitalize";

import Typography from "@material-ui/core/Typography";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemText from "@material-ui/core/ListItemText";

import ButtonSet from "/components/ButtonSet";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import StatusMenu from "/components/partials/StatusMenu";
import ListHeader from "/components/partials/ListHeader";
import ButtonMenu from "/components/partials/ButtonMenu";

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

class EventsListHeader extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    onClickSelect: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
    onClickResolve: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    selectedCount: PropTypes.number.isRequired,
    rowCount: PropTypes.number.isRequired,
    environment: PropTypes.shape({
      checks: PropTypes.object,
      entities: PropTypes.object,
    }),
    onChangeQuery: PropTypes.func.isRequired,
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
    this.props.onChangeQuery({ filter: `Entity.ID == '${newValue}'` });
  };

  requeryCheck = newValue => {
    this.props.onChangeQuery({ filter: `Check.Name == '${newValue}'` });
  };

  requeryStatus = newValue => {
    if (Array.isArray(newValue)) {
      if (newValue.length === 1) {
        this.props.onChangeQuery({ filter: `Check.Status == ${newValue}` });
      } else {
        const val = newValue.join(",");
        this.props.onChangeQuery({ filter: `Check.Status IN (${val})` });
      }
    } else {
      this.props.onChangeQuery(query => query.delete("filter"));
    }
  };

  _handleChangeSort = newValue => {
    this.props.onChangeQuery({ order: newValue });
  };

  render() {
    const {
      classes,
      selectedCount,
      rowCount,
      onClickSelect,
      onClickSilence,
      onClickResolve,
      onClickDelete,
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
      <ListHeader
        sticky
        selectedCount={selectedCount}
        rowCount={rowCount}
        onClickSelect={onClickSelect}
        renderBulkActions={() => (
          <ButtonSet>
            <ConfirmDelete
              identifier={`${selectedCount} ${
                selectedCount === 1 ? "event" : "events"
              }`}
              onSubmit={onClickDelete}
            >
              {confirm => (
                <Button onClick={confirm.open}>
                  <Typography variant="button">Delete</Typography>
                </Button>
              )}
            </ConfirmDelete>
            <Button onClick={onClickSilence}>
              <Typography variant="button">Silence</Typography>
            </Button>
            <Button onClick={onClickResolve}>
              <Typography variant="button">Resolve</Typography>
            </Button>
          </ButtonSet>
        )}
        renderActions={() => (
          <ButtonSet>
            <ButtonMenu label="Entity" onChange={this.requeryEntity}>
              {entityNames.map(name => (
                <MenuItem key={name} value={name}>
                  <ListItemText primary={name} />
                </MenuItem>
              ))}
            </ButtonMenu>
            <ButtonMenu label="Check" onChange={this.requeryCheck}>
              {checkNames.map(name => (
                <MenuItem key={name} value={name}>
                  <ListItemText primary={name} />
                </MenuItem>
              ))}
            </ButtonMenu>
            <StatusMenu onChange={this.requeryStatus} />
            <ButtonMenu
              className={classes.headerButton}
              label="Sort"
              onChange={this._handleChangeSort}
            >
              {["SEVERITY", "NEWEST", "OLDEST"].map(name => (
                <MenuItem key={name} value={name}>
                  <ListItemText primary={capitalize(name)} />
                </MenuItem>
              ))}
            </ButtonMenu>
          </ButtonSet>
        )}
      />
    );
  }
}

export default withStyles(styles)(EventsListHeader);
