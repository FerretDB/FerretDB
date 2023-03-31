import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Easy to Use',
    Svg: require('@site/static/img/undraw-ferretdb-usability.svg').default,
    description: (
      <>
        Easy to use document databases that protects you from vendor lock-in and fauxpen licenses.
      </>
    ),
  },
  {
    title: 'Open-Source',
    Svg: require('@site/static/img/undraw-ferretdb-open-source.svg').default,
    description: (
      <>
        Perfect open-source software for those looking for MongoDB development experience.
      </>
    ),
  },
  {
    title: 'MongoDB Alternative',
    Svg: require('@site/static/img/undraw-ferretdb-server.svg').default,
    description: (
      <>
        Compatible with MongoDB drivers and should work as a drop-in replacement to MongoDB in many cases.
      </>
    ),
  },
];

function Feature({Svg, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
