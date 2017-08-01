# String metrics

This library contains implementations of the Levenshtein distance, Jaro-Winkler and Soundex algorithms written in Go (golang). Other algorithms related with string metrics (or string similarity, whatever) are welcome.

* master: [![Build Status](https://travis-ci.org/xrash/smetrics.svg?branch=master)](http://travis-ci.org/xrash/smetrics)

# Algorithms

## WagnerFischer

        func WagnerFischer(a, b string, icost, dcost, scost int) int

The Wagner-Fischer algorithm for calculating the Levenshtein distance. It runs on O(mn) and needs O(2m) space where m is the size of the smallest string. This is kinda optimized so it should be used in most cases.

The first two parameters are the two strings to be compared. The last three parameters are the insertion cost, the deletion cost and the substitution cost. These are normally defined as 1, 1 and 2.

#### Examples:

        smetrics.WagnerFischer("POTATO", "POTATTO", 1, 1, 2)
		>> 1, delete the second T on POTATTO

        smetrics.WagnerFischer("MOUSE", "HOUSE", 2, 2, 4)
		>> 4, substitute M for H

## Ukkonen

        func Ukkonen(a, b string, icost, dcost, scost int) int

The Ukkonen algorithm for calculating the Levenshtein distance. The algorithm is described [here](http://www.cs.helsinki.fi/u/ukkonen/InfCont85.PDF). It runs on O(t . min(m, n)) where t is the actual distance between strings a and b, so this version should be preferred over the WagnerFischer for strings **very** similar. In practice, it's slower most of the times. It needs O(min(t, m, n)) space.

The first two parameters are the two strings to be compared. The last three parameters are the insertion cost, the deletion cost and the substitution cost. These are normally defined as 1, 1 and 2.

#### Examples:

        smetrics.Ukkonen("POTATO", "POTATTO", 1, 1, 2)
		>> 1, delete the second T on POTATTO

        smetrics.Ukkonen("MOUSE", "HOUSE", 2, 2, 4)
		>> 4, substitute M for H

## Jaro

        func Jaro(a, b string) float64

The Jaro distance. It is not very accurate, therefore you should prefer the JaroWinkler optimized version.

#### Examples:

        smetrics.Jaro("AL", "AL")
		>> 1, equal strings

        smetrics.Jaro("MARTHA", "MARHTA")
		>> 0.9444444444444445, very likely a typo

        smetrics.Jaro("JONES", "JOHNSON")
		>> 0.7904761904761904

## JaroWinkler

        func JaroWinkler(a, b string, boostThreshold float64, prefixSize int) float64

The JaroWinkler distance. JaroWinkler returns a number between 0 and 1 where 1 means perfectly equal and 0 means completely different. It is commonly used on Record Linkage stuff, thus it tries to be accurate for real names and common typos. You should consider it on data such as person names and street names.

JaroWinkler is a more accurate version of the Jaro algorithm. It works by boosting the score of exact matches at the beginning of the strings. By doing this, Winkler says that typos are less common to happen at the beginning. For this to happen, it introduces two more parameters: the boostThreshold and the prefixSize. These are commonly set to 0.7 and 4, respectively.

#### Examples:

        smetrics.JaroWinkler("AL", "AL", 0.7, 4)
		>> 1, equal strings

        smetrics.JaroWinkler("MARTHA", "MARHTA", 0.7, 4)
		>> 0.9611111111111111, very likely a typo

        smetrics.JaroWinkler("JONES", "JOHNSON", 0.7, 4)
		>> 0.8323809523809523

## Soundex

        func Soundex(s string) string

The Soundex encoding. It is a phonetic algorithm that considers how the words sound in english. Soundex maps a name to a 4-byte string consisting of the first letter of the original string and three numbers. Strings that sound similar should map to the same thing.

#### Examples:

        smetrics.Soundex("Euler")
		>> E460

        smetrics.Soundex("Ellery")
		>> E460

        smetrics.Soundex("Lloyd")
		>> L300

        smetrics.Soundex("Ladd")
		>> L300

## Hamming

        func Hamming(a, b string) (int, error)

The Hamming distance is simply the minimum number of substitutions required to change one string into the other. Both strings must have the same size, of the function returns an error.

#### Examples:

        smetrics.Hamming("aaa", "aaa")
		>> 0, nil

        smetrics.Hamming("aaa", "aab")
		>> 1, nil

        smetrics.Hamming("aaaa", "a")
		>> -1, error

# TODO

- Accept cost functions instead of constant values in every Levenshtein implementation.

- Make a better interface.

- Moar algos!
