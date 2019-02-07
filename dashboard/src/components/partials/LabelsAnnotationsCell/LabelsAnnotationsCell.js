import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import CardContent from "@material-ui/core/CardContent";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/components/Dictionary";
import Grid from "@material-ui/core/Grid";
import Maybe from "/components/Maybe";
import Label from "/components/partials/Label";

class LabelsAnnotationsCell extends React.PureComponent {
  static propTypes = {
    entity: PropTypes.object,
    check: PropTypes.object,
  };

  static defaultProps = {
    entity: null,
    check: null,
  };

  static fragments = {
    check: gql`
      fragment LabelsAnnotationsCell_check on CheckConfig {
        metadata {
          labels {
            key
            val
          }
          annotations {
            key
            val
          }
        }
      }
    `,
    entity: gql`
      fragment LabelsAnnotationsCell_entity on Entity {
        metadata {
          labels {
            key
            val
          }
          annotations {
            key
            val
          }
        }
      }
    `,
  };

  render() {
    const { check, entity } = this.props;

    const object = check || entity;

    return (
      <CardContent>
        <Grid container spacing={0}>
          <Grid item xs={12} sm={6}>
            <Dictionary>
              <DictionaryEntry>
                <DictionaryKey>Labels</DictionaryKey>
                <DictionaryValue>
                  <Maybe value={object.metadata.labels} fallback="None">
                    {val =>
                      val.map(pair => [
                        <Label name={pair.key} value={pair.val} />,
                        " ",
                      ])
                    }
                  </Maybe>
                </DictionaryValue>
              </DictionaryEntry>
              <DictionaryEntry>
                <DictionaryKey>Annotations</DictionaryKey>
                <DictionaryValue>
                  <Maybe value={object.metadata.annotations} fallback="None">
                    {val =>
                      val.map(pair => [
                        <Label name={pair.key} value={pair.val} />,
                        " ",
                      ])
                    }
                  </Maybe>
                </DictionaryValue>
              </DictionaryEntry>
            </Dictionary>
          </Grid>
        </Grid>
      </CardContent>
    );
  }
}

export default LabelsAnnotationsCell;
